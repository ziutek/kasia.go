include $(GOROOT)/src/Make.inc

TARG=kasia
GOFILES=\
	parser_elements.go\
	parser1.go\
	parser2.go\
	parser_err.go\
	rev_parser.go\
	getvarfun.go\
	template.go\
	compat.go\

include $(GOROOT)/src/Make.pkg

