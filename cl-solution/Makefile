LISP ?= sbcl

build:
	$(LISP) --load shop-tasks.asd \
		--eval '(ql:quickload :shop-tasks)' \
		--eval '(asdf:make :shop-tasks)' \
		--eval '(quit)'
