run:
	go run main/main.go

install:
	go build -o ~/go/bin/simplefs main/main.go 

update:
	git add . && git commit -m "update" && git push