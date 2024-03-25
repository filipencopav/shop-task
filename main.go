package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

func printAsJson(data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil { return err
	} else { fmt.Print(string(b)); return nil }
}

func main() {
	strs := strings.Split(os.Args[1], ",")
	orders := make([]int, len(strs))
	for i, v := range strs {
		var err error
		orders[i], err = strconv.Atoi(v)
		exitOnErr(err)
	}

	// Can also be printed as Json. See `printAsJson`.
	// It's just a bunch of maps
	data := getStructuredData(orders)

	fmt.Println("=+=+=+=")
	fmt.Printf("Страница сборки заказов %s\n\n", os.Args[1])
	for _, shelf := range data {
		if len(shelf["commodities"].([]map[string]any)) > 0 {
			fmt.Println("===Стеллаж", shelf["title"])
		}
		for _, commodity := range shelf["commodities"].([]map[string]any) {
			fmt.Printf("%s (id=%d)\n",
				commodity["title"],
				commodity["id"])
			fmt.Printf("заказ %d, %d шт\n",
				commodity["order_id"],
				commodity["quantity"])
			if len(commodity["otherShelves"].([]int32)) > 0 {
				fmt.Printf("доп стеллаж: ")
				for i, index := range commodity["otherShelves"].([]int32) {
					if i > 0 {
						fmt.Printf(",")
					}
					fmt.Printf("%s", data[index]["title"])
				}
				fmt.Printf("\n")
			}
			fmt.Printf("\n")
		}
	}
}

const orders_q = `SELECT id, commodity, quantity
  FROM orders
  WHERE id = ANY($1::int[])
  ORDER BY (id) ASC`
const shelves_q = `SELECT DISTINCT commodity, shelf
  FROM commodities_shelves
  WHERE commodity = ANY($1::int[])
  ORDER BY (shelf) ASC`
const info_q = `SELECT id, title, main_shelf
  FROM commodities WHERE id = ANY($1::int[])`
const shelf_title_q = `SELECT id, title
  FROM shelves
  WHERE id = ANY($1::int[])`

func getStructuredData(orders_list []int) map[int32]map[string]any {
	// 1. orders = query(orders_q, orders_list)
	//     => [{order_id: 10, comm: 2, quant: 3} ...]
	orders := make([]map[string]int32, 0)
	iterQuery(
		orders_q,
		func(values []any) {
			orders = append(orders, map[string]int32{
				"order_id": values[0].(int32),
				"comm":     values[1].(int32),
				"quant":    values[2].(int32),
			})
		},
		orders_list,
	)

	// 2. pairs = query(shelves_q, orders.fmap(order -> order.comm))
	//   => [{comm: 2, shelf: 1},
	//       {comm: 2, shelf: 3}, ...]
	//    'pairs' will contain secondary shelves, thus duplicates of comm ids
	// 4. Group by commodity id: pairs = {2 [1, 3], ...}
	pairs := make(map[int32]map[string]any, 0)
	allIncludedShelves := make(map[int32]string)
	iterQuery(shelves_q,
		func(values []any) {
			comm := values[0].(int32)
			shelf := values[1].(int32)
			allIncludedShelves[shelf] = "" // We will need this way down
			if pairs[comm] == nil {
				pairs[comm] = map[string]any{
					"shelves": make([]int32, 0),
				}
			}
			pairs[comm]["shelves"] = append(
				pairs[comm]["shelves"].([]int32),
				shelf,
			)
		},
		fmap(func(order map[string]int32) int32 {
			return order["comm"]
		},
			orders),
	)

	// 5. info_q -> create grouped maps, get shelf data from comm_shelves
	//    split secondary shelves from main shelf
	//    {
	//      2: {
	//        title: "..",
	//        main_shelf: 1,
	//        secondary_shelves: [3]
	//      },
	//    }
	iterQuery(info_q,
		func(values []any) {
			id, title, main_shelf :=
				values[0].(int32),
				values[1].(string),
				values[2].(int32)

			pairs[id]["title"] = title
			pairs[id]["main_shelf"] = main_shelf
			pairs[id]["secondary_shelves"] = filter(func(shelf int32) bool {
				return shelf != main_shelf
			},
				pairs[id]["shelves"].([]int32))
		},
		fmap(func(order map[string]int32) int32 {
			return order["comm"]
		}, orders),
	)

	// 6. Group by order, by iterating thru 'orders' and adding to new map:
	//    {10  //  order id (key in a map)
	//     [{:quant 3
	//       :id 2
	//       :title "..",
	//       :main_shelf 1,
	//       :secondary_shelves [3]
	//      },
	//      ...}],
	//    ...}
	ordersGroup := make(map[int32][]map[string]any)
	for _, m := range orders {
		val := ordersGroup[m["order_id"]]
		if val == nil {
			val = make([]map[string]any, 0)
		}
		val = append(val, map[string]any{
			"id":                m["comm"],
			"quantity":          m["quant"],
			"title":             pairs[m["comm"]]["title"],
			"main_shelf":        pairs[m["comm"]]["main_shelf"],
			"secondary_shelves": pairs[m["comm"]]["secondary_shelves"],
			"rest":              pairs[m["comm"]],
		})
		ordersGroup[m["order_id"]] = val
	}

	shelves := make(map[int32]map[string]any)
	
	keys := make([]int32, len(allIncludedShelves))
	i := 0
	for k := range allIncludedShelves {
		keys[i] = k
		i++
	}
	iterQuery(shelf_title_q,
		func(values []any) {
			id := values[0].(int32)
			title := values[1].(string)
			if shelves[id] == nil {
				shelves[id] = map[string]any {
					"commodities": make([]map[string]any, 0),
				}
			}
			shelves[id]["title"] = title
		},
		keys)

	// 7. Extract main shelves to top level and restructure commodities:
	//    {1  //  shelf
	//     {:title "..."
	//      :commodities [{:id 2
	//                     :order 10
	//                     :title "..",
	//                     :quant 3, etc.}]}}
	for order_id, order_comms := range ordersGroup {
		for _, comm := range order_comms {
			main_shelf := comm["main_shelf"].(int32)
			final_comm := map[string]any {
				"id":           comm["id"],
				"order_id":     order_id,
				"quantity":     comm["quantity"],
				"otherShelves": comm["secondary_shelves"].([]int32),
				"title":        comm["title"],
			}

			shelves[main_shelf]["commodities"] = append(
				shelves[main_shelf]["commodities"].([]map[string]any),
				final_comm,
			)
			// we will init shelves[main_shelf]["title"] later
		}
	}

	return shelves
}

func exitOnErr(err error) {
	if err != nil {
		_, filename, line, _ := runtime.Caller(1)
		_, f2, l2, _ := runtime.Caller(2)
		_, f3, l3, _ := runtime.Caller(3)
		_, f4, l4, _ := runtime.Caller(4)
		fmt.Fprintf(os.Stderr, ":: err: %s:%d\n>> %v\n", filename, line, err)
		fmt.Fprintf(os.Stderr, "   IN: %s:%d\n", f2, l2)
		fmt.Fprintf(os.Stderr, "   IN: %s:%d\n", f3, l3)
		fmt.Fprintf(os.Stderr, "   IN: %s:%d\n", f4, l4)
		os.Exit(1)
	}
}

func iterQuery(query string, fn func(values []any), args ...any) {
	db_url := "postgresql://shop-tasks@localhost"

	conn, err := pgx.Connect(context.Background(), db_url)
	exitOnErr(err)
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), query, args...)
	exitOnErr(err)
	defer rows.Close()

	for rows.Next() {
		values, err := rows.Values()
		exitOnErr(err)

		fn(values)
	}
	exitOnErr(rows.Err())
}

func fmap[A, B any](fn func(A) B, array []A) []B {
	res := make([]B, len(array))
	for i := range array {
		res[i] = fn(array[i])
	}
	return res
}

func filter[A any](fn func(A) bool, array []A) []A {
	res := make([]A, 0)
	for i := range array {
		if fn(array[i]) {
			res = append(res, array[i])
		}
	}
	return res
}

