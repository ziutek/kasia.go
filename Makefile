include $(GOROOT)/src/Make.inc

TARG=github.com/ziutek/kasia.go
GOFILES=\
	parser_elements.go\
	parser1.go\
	parser2.go\
	parser_err.go\
	getvarfun.go\
	template.go\
	compat.go\
#	rev_parser.go\

include $(GOROOT)/src/Make.pkg

