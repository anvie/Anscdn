include $(GOROOT)/src/Make.inc

CC=$(GOBIN)/6g 
LD=$(GOBIN)/6l 

OBJS =\
	anlog.6 \
	configfile.6 \
	filemon.6 \
	config.6 \
	utils.6 \
	downloader.6 \
	cdnize.6 \
	anscdn.6

TESTSSRC = \
	cdnize_test.go

all: anscdn

anscdn: $(OBJS)
	$(LD) -o $@ anscdn.$(O)

%.6 : %.go 
	$(GC) $< 

% : %.6 
	$(LD) -L . -o $@ $^ 

clean:
	rm -f *.$(O)
	