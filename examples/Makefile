include $(GOROOT)/src/Make.inc

ALL=benchmarking simple long_time pointers  webgo_nested webgo_ctxstack

all: $(ALL)

clean:
	rm -rf *.[68] $(ALL)

%: %.go
	$(GC) $*.go
	$(LD) -o $@ $@.$O
