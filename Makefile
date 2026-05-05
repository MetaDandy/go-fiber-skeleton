.PHONY: generate
generate:
	gorm generate -i ./src/model -o ./src/generated
