package main

import (
	"context"
	"go-fun/ent"
	"go-fun/ent/todo"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")

	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema: %v", err)
	}

	task1, err := client.Todo.Create().SetText("Add GraphQL example").Save(ctx)
	if err != nil {
		log.Fatalf("failed creating todo: %v", err)
	}

	log.Printf("Created task 1: %+v", task1)

	task2, err := client.Todo.Create().SetText("Add Tracing Example").Save(ctx)
	if err != nil {
		log.Fatalf("failed creating todo: %v", err)
	}

	log.Printf("Created task 2: %+v", task2)

	if err := task2.Update().SetParent(task1).Exec(ctx); err != nil {
		log.Fatalf("failed updating todo: %v", err)
	}

	items, err := client.Todo.Query().All(ctx)
	if err != nil {
		log.Fatalf("failed querying todos: %v", err)
	}
	for _, item := range items {
		log.Printf("Todo: %+v", item)
	}

	log.Println("Dependencies:")
	items, err = client.Todo.Query().Where(todo.HasParent()).All(ctx)
	if err != nil {
		log.Fatalf("failed querying todos: %v", err)
	}
	for _, item := range items {
		log.Printf("dependency: %+v", item)
	}
}
