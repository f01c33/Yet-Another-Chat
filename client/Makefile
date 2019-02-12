CC = go
# gcc
# icc
# tcc
# SRC = server
# EXT = .go
SRC2 = client
EXT2 = .go
# .c
# .cpp
C_FLAGS = 
#-Wall -g -fPIC -std=gnu11
#-std=c99
INC_PATH = 
# -I/path/to/include/files
LIB_PATH = 
# -L/usr/lib/ -L/path/to/libs 
LIBS = 
#-lm

all: ${SRC2}

# ${SRC}: #${SRC}.o
# 	$(CC) build ${SRC}${EXT} 
#-o ${SRC2} $(LIB_PATH) $(LIBS)

${SRC2}: #${SRC2}.o
	$(CC) build ${SRC2}${EXT2}
#-o ${SRC2} $(LIB_PATH) $(LIBS)
#	./${SRC}

# ${SRC}.o: ${SRC}${EXT}
# 	$(CC) -c ${SRC}${EXT} ${C_FLAGS}

# ${SRC2}.o: ${SRC2}${EXT2}
# 	$(CC) -c ${SRC2}${EXT2} ${C_FLAGS}

clean:
	# rm ${SRC}
	rm ${SRC2}