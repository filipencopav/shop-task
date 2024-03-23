(asdf:defsystem shop-tasks
  :depends-on (#:postmodern #:alexandria #:group-by #:arrows #:split-sequence)
  :components ((:file "main"))
  :build-operation "program-op"
  :build-pathname "fetch-orders"
  :entry-point "shop-tasks:main")
