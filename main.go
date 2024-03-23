package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

const query = `SELECT
    shelves.title AS stellaj,
    shelves.id AS stellaj_id,
    commodities.title AS tovar,
    orders.commodity AS tovar_id,
    orders.id AS order_id,
    orders.quantity AS kolichestvo,
    ARRAY(SELECT DISTINCT sh.title
          FROM shelves as sh
          JOIN commodities_shelves ON commodities_shelves.shelf = sh.id
               AND commodities_shelves.commodity = commodities.id
          WHERE NOT is_main_shelf
          ORDER BY (sh.title) DESC)::text[]
      AS dop_stellajy
FROM orders
JOIN commodities ON commodities.id = orders.commodity
JOIN commodities_shelves ON commodities_shelves.commodity = commodities.id
JOIN shelves ON commodities_shelves.shelf = shelves.id
WHERE commodities_shelves.is_main_shelf
      AND orders.id = ANY($1::integer[])
ORDER BY (shelves.title, orders.id) ASC;`

type Order struct {
	commodity_title    string
	commodity_id       int32
	order_id           int32
	count              int32
	additional_shelves []interface{}
}

type Shelf struct {
	title  string
	id     int32
	orders []Order
}

func get_data(rows pgx.Rows) []Shelf {
	shelves := []Shelf{}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR]: %v\n", err)
			os.Exit(1)
		}

		title := values[0].(string)
		id := values[1].(int32)

		shelf := &Shelf{}
		found := false
		for i := 0; i < len(shelves); i++ {
			s := &shelves[i]
			if s.id == id {
				shelf = s
				found = true
			}
		}

		shelf.id = id
		shelf.title = title
		order := Order{
			commodity_title:    values[2].(string),
			commodity_id:       values[3].(int32),
			order_id:           values[4].(int32),
			count:              values[5].(int32),
			additional_shelves: values[6].([]interface{}),
		}
		shelf.orders = append(shelf.orders, order)

		if !found {
			shelves = append(shelves, *shelf)
		}
	}

	err := rows.Err()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при извлечении данных: %v\n", err)
		os.Exit(1)
	}

	return shelves
}

func get_rows(orders []int) pgx.Rows {
	db_url := "postgresql://shop-tasks@localhost"

	conn, err := pgx.Connect(context.Background(), db_url)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Не получилось подключиться к базе данных: %v\n",
			err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(context.Background(), query, orders)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Ошибка запроса: %v\n",
			err)
		os.Exit(1)
	}

	return rows
}

func main() {
	orders_strs := strings.Split(os.Args[1], ",")
	orders := make([]int, len(orders_strs))
	for i, v := range orders_strs {
		var err error
		orders[i], err = strconv.Atoi(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка извлечения номеров заказов: %v", err)
			os.Exit(1)
		}
	}
	shelves := get_data(get_rows(orders))
	fmt.Println("=+=+=+=")
	fmt.Printf("Страница сборки заказов %s\n\n", os.Args[1])
	for _, shelf := range shelves {
		fmt.Println("===Стеллаж", shelf.title)
		for _, order := range shelf.orders {
			fmt.Printf("%s (id=%d)\n",
				order.commodity_title,
				order.commodity_id)
			fmt.Printf("заказ %d, %d шт\n", order.order_id, order.count)
			if len(order.additional_shelves) > 0 {
				fmt.Printf("доп стеллаж: ")
				for i, s := range order.additional_shelves {
					if i > 0 {
						fmt.Printf(",")
					}
					fmt.Printf("%s", s)
				}
				fmt.Printf("\n")
			}
			fmt.Printf("\n")
		}
	}
}
