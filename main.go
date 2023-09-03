package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

type TODOId uint

type Todo struct {
	id        TODOId
	completed bool
	title     string
}

func (todo *Todo) complete() {
	todo.completed = true
}

func (todo *Todo) uncomplete() {
	todo.completed = false
}

func (todo *Todo) update(title string) {
	todo.title = title
}

func (todo Todo) save() {
	filepath := getFilePath(todo.id)
	file, err := os.Create(filepath)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	completedByte := make([]byte, 1)
	if todo.completed {
		completedByte[0] = 0x1
	} else {
		completedByte[0] = 0x0
	}
	file.Write(completedByte)

	titleBytes := []byte(todo.title)
	file.Write(titleBytes)
}

func (todo Todo) delete() {
	filepath := getFilePath(todo.id)
	err := os.Remove(filepath)
	if err != nil {
		fmt.Printf("Error deleting todo file %s: %+v\n", filepath, err)
	}
}

func (todo Todo) print() {
	fmt.Printf("%d\t%s\n", todo.id, todo.title)
}

func printHelp() {
	fmt.Print(
		"Select action:\n" +
			"1: List all TODOs\n" +
			"2: Add new TODO\n" +
			"3: Complete TODO\n" +
			"4: Uncomplete TODO\n" +
			"5: Delete TODO\n" +
			"6: List completed TODOs\n" +
			"7: List uncompleted TODOs\n" +
			"8: Edit TODO\n" +
			"9: Show this help\n" +
			"0: Exit\n",
	)
}

func getTodoId() TODOId {
	for {
		var idRaw string
		fmt.Print("Select todo: ")
		fmt.Scanln(&idRaw)

		idInt, err := strconv.Atoi(idRaw)
		if err != nil {
			fmt.Printf("Error reading id: %+v\n", err)
			continue
		}

		id := TODOId(idInt)
		return id
	}
}

func getTodoTitle() string {
	in := bufio.NewReader(os.Stdin)
	title, _ := in.ReadString('\n')
	return strings.Trim(title, "\n")
}

func getDirPath() string {
	dir, err := os.Getwd()

	if err != nil {
		log.Fatal(err)
	}

	return path.Join(dir, "todos")
}

func getFilePath(id TODOId) string {
	dir := getDirPath()
	filepath := path.Join(dir, strconv.Itoa(int(id)))

	return filepath
}

func LoadTodo(id TODOId) (*Todo, error) {
	filepath := getFilePath(id)
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	completedByte := make([]byte, 1)
	data := make([]byte, 64)

	// I couldn't find a way to read only first bit, so reading the first byte and checking it's value
	file.Read(completedByte)
	completed := completedByte[0]&0x1 == 0x1

	title := ""
	for {
		n, err := file.Read(data)
		if err == io.EOF {
			break
		}
		title += string(data[:n])
	}

	return &Todo{
		id:        id,
		completed: completed,
		title:     title,
	}, nil
}

func createTodoItem() {
	fmt.Print("title: ")
	title := getTodoTitle()

	todo := &Todo{
		id:        TODOId(rand.Intn(1000)),
		title:     title,
		completed: false,
	}
	todo.save()

	fmt.Printf("Saved with id: %d\n", todo.id)
}

func changeTodoItemState(complete bool) {
	id := getTodoId()
	todo, err := LoadTodo(id)

	if err != nil {
		fmt.Println("Todo not found")
		return
	}

	if complete {
		todo.complete()
	} else {
		todo.uncomplete()
	}

	todo.save()

	fmt.Println("Todo updated")
}

func editTodo() {
	id := getTodoId()
	todo, err := LoadTodo(id)

	if err != nil {
		fmt.Println("Todo not found")
		return
	}

	fmt.Printf("old title: %s", todo.title)
	fmt.Print("new title: ")
	title := getTodoTitle()

	todo.update(title)
	todo.save()

	fmt.Println("Todo updated")
}

func deleteTodo() {
	id := getTodoId()
	todo, err := LoadTodo(id)

	if err != nil {
		fmt.Println("Todo not found")
		return
	}

	todo.delete()

	fmt.Println("Todo deleted")
}

func loadAllTodos() []*Todo {
	dir := getDirPath()
	entries, err := os.ReadDir(dir)

	if err != nil {
		log.Fatal(err)
	}

	todos := make([]*Todo, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// can check using strconv, but I wanted to try regexp
		match, _ := regexp.MatchString("^([0-9]+)$", entry.Name())
		if !match {
			continue
		}

		idInt, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		id := TODOId(idInt)
		todo, err := LoadTodo(id)

		if err != nil {
			continue
		}

		todos = append(todos, todo)
	}

	return todos
}

func listTodos(includeUncomplete, includeComplete bool) {
	todos := loadAllTodos()

	completedTodos := make([]*Todo, 0, len(todos))
	uncompletedTodos := make([]*Todo, 0, len(todos))

	for _, todo := range todos {
		if todo.completed {
			completedTodos = append(completedTodos, todo)
		} else {
			uncompletedTodos = append(uncompletedTodos, todo)
		}
	}

	if includeUncomplete {
		fmt.Printf("%d uncompleted todos:\n", len(uncompletedTodos))
		for _, todo := range uncompletedTodos {
			todo.print()
		}
	}

	if includeUncomplete && includeComplete {
		fmt.Print("\n")
	}

	if includeComplete {
		fmt.Printf("%d completed todos:\n", len(completedTodos))
		for _, todo := range completedTodos {
			todo.print()
		}
	}
}

func main() {
	fmt.Println("Simple CLI TODO app")

	printHelp()

	for {
		var actionRaw string
		fmt.Print("> ")
		fmt.Scanln(&actionRaw)
		action, err := strconv.Atoi(actionRaw)

		if err != nil {
			fmt.Printf("Error reading action: %+v\n", err)
			continue
		}

		switch action {
		case 1:
			listTodos(true, true)
		case 2:
			createTodoItem()
		case 3:
			changeTodoItemState(true)
		case 4:
			changeTodoItemState(false)
		case 5:
			deleteTodo()
		case 6:
			listTodos(true, false)
		case 7:
			listTodos(false, true)
		case 8:
			editTodo()
		case 9:
			printHelp()
		case 0:
			fmt.Println("Goodbye!")
			os.Exit(0)
		default:
			fmt.Println("Unknown action")
		}
	}
}
