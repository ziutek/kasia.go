include $(GOROOT)/src/Make.inc

ALL=simple webgo_nested webgo_ctxstack long_time

all: $(ALL)

clean:
	rm -rf *.[68] $(ALL)

%: %.go
	$(GC) $*.go
	$(LD) -o $@ $@.$O
