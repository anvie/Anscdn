CC=$(GOBIN)/6g 
LD=$(GOBIN)/6l 

OBJS =\
	anlog.6 \
	configfile.6 \
	filemon.6 \
	config.6 \
	utils.6 \
	anscdn.6

all: anscdn

anscdn: $(OBJS)
	$(LD) -o $@ anscdn.6

%.6 : %.go 
	$(CC) $< 

% : %.6 
	$(LD) -L . -o $@ $^ 

clean:
	rm *.6