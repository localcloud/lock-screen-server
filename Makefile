OUT_DIR = builds
BIN_NAME = app

build:
	govendor build -o $(OUT_DIR)/$(BIN_NAME)
dep:
	govendor sync
run: dep build
	chmod +x $(OUT_DIR)/$(BIN_NAME); ./$(OUT_DIR)/$(BIN_NAME) --port=8080 --db=/tmp/lock-screen-saver.db.json