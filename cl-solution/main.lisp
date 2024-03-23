(uiop:define-package #:shop-tasks
  (:use :cl :arrows :group-by :split-sequence)
  (:export #:main))

(in-package #:shop-tasks)

(defparameter query
  "SELECT
    shelves.id AS stellaj_id,
    shelves.title AS stellaj,
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
ORDER BY (shelves.title, orders.id) ASC;")

(defmacro comment (&body body)
  (declare (ignore body))
  nil)

(defmacro assoc-get (alist key)
  `(cdr (assoc ,key ,alist)))

(comment
  (pomo:with-connection '("shop-tasks" "shop-tasks" "" "localhost")
    (-> (pomo:query query '(10 11 14 15) :alists)
        (group-by)))
  )

(defun get-grouped-data (orders)
  (pomo:with-connection '("shop-tasks" "shop-tasks" "" "localhost")
    (-> (pomo:query query orders :alists)
        (group-by
         :key (lambda (list) (list (first list) (second list) ()))
         :value #'cddr))))

(defun main ()
  (format t "=+=+=+=~%")
  (format t "Страница сборки заказов~%~%")
  (let* ((args (-> (uiop:command-line-arguments)
                   (first)
                   (or (error "Must supply one argument, a comma-separated list of order ids"))
                   (->> (split-sequence #\,))))
         (data (get-grouped-data args)))
    (mapc
     (lambda (shelf)
       (format t "===Стеллаж ~A~%" (cdr (assoc :stellaj (car shelf))))
       (mapc
        (lambda (commodity)
          (format t "~A (id=~D)~%заказ ~D, ~D шт~%"
                  (assoc-get commodity :tovar)
                  (assoc-get commodity :tovar-id)
                  (assoc-get commodity :order-id)
                  (assoc-get commodity :kolichestvo))
          (let ((additional-shelves (assoc-get commodity :dop-stellajy)))
            (unless (uiop:emptyp additional-shelves)
              (format t "доп стеллаж: ~{~D~^,~}"
                      (coerce additional-shelves 'list))))
          (format t "~&~%"))
        (cdr shelf)))
     data)))
