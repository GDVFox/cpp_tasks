package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

//CommandID представляется индекс комманды, используется для map
type CommandID int

//Command структура представляющая команду
type Command struct {
	ID     CommandID
	TimeIn int

	Ancestor  *Command
	Label     *Command
	Sdom      *Command
	Idom      *Command
	DFSParent *Command

	Nexts   []*Command
	Parents []*Command
	Bucket  []*Command
}

//CommandSlice нужен для сортировки на версии ниже 1.8
type CommandSlice []*Command

func (cs CommandSlice) Len() int {
	return len(cs)
}

func (cs CommandSlice) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

func (cs CommandSlice) Less(i, j int) bool {
	return cs[i].TimeIn > cs[j].TimeIn
}

var globalDFSTime = 1

func dfs(current *Command) {
	current.TimeIn = globalDFSTime
	globalDFSTime++

	for _, next := range current.Nexts {
		if next.TimeIn == 0 {
			next.DFSParent = current
			dfs(next)
		}
	}
}

func findMin(v *Command) *Command {
	searchAndCut(v)
	return v.Label
}

func searchAndCut(v *Command) *Command {
	if v.Ancestor == nil {
		return v
	}

	res := searchAndCut(v.Ancestor)
	if v.Ancestor.Label.Sdom.TimeIn < v.Label.Sdom.TimeIn {
		v.Label = v.Ancestor.Label
	}

	v.Ancestor = res
	return res
}

func dominators(root *Command, commands []*Command) {
	for _, command := range commands {
		command.Sdom = command
		command.Label = command
	}

	var rootPos int
	for i, current := range commands {
		if current == root {
			rootPos = i
			break
		}

		for _, v := range current.Parents {
			if u := findMin(v); u.TimeIn != 0 &&
				u.Sdom.TimeIn < current.Sdom.TimeIn {
				current.Sdom = u.Sdom
			}
		}
		current.Ancestor = current.DFSParent
		current.Sdom.Bucket = append(current.Sdom.Bucket, current)

		for _, v := range current.DFSParent.Bucket {
			if u := findMin(v); u.Sdom == v.Sdom {
				v.Idom = v.Sdom
			} else {
				v.Idom = u
			}
		}

		current.DFSParent.Bucket = current.DFSParent.Bucket[:0]
	}

	for i := rootPos - 1; i >= 0; i-- {
		current := commands[i]

		if current.Idom != current.Sdom {
			current.Idom = current.Idom.Idom
		}
	}

	root.Idom = nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	var n int
	var err error
	if scanner.Scan() {
		if n, err = strconv.Atoi(scanner.Text()); err != nil {
			log.Fatalf("reading error: %v", err)
		}
	} else {
		log.Fatal("expected number of commands")
	}

	var firstCommand, prevCommand *Command
	idToCommandIndex := make(map[CommandID]int, n)
	commands := make([]*Command, 0, n)
	for scanner.Scan() {
		commandsComponents := strings.Split(scanner.Text(), " ")
		id, err := strconv.Atoi(commandsComponents[0])
		if err != nil {
			log.Fatalf("incorrect command ID %s: %v", commandsComponents[0], err)
		}

		var newCommand *Command
		if commandIndex, ok := idToCommandIndex[CommandID(id)]; !ok {
			newCommand = &Command{
				ID:      CommandID(id),
				Bucket:  make([]*Command, 0),
				Parents: make([]*Command, 0),
				Nexts:   make([]*Command, 0),
			}

			if len(commands) == 0 {
				firstCommand = newCommand
			}

			commands = append(commands, newCommand)
			idToCommandIndex[CommandID(id)] = len(commands) - 1
		} else {
			newCommand = commands[commandIndex]
		}

		if prevCommand != nil {
			prevCommand.Nexts = append(prevCommand.Nexts, newCommand)
			newCommand.Parents = append(newCommand.Parents, prevCommand)
		}

		switch commandsComponents[1] {
		case "ACTION":
			prevCommand = newCommand
		case "BRANCH", "JUMP":
			to, err := strconv.Atoi(commandsComponents[2])
			if err != nil {
				log.Fatalf("incorrect command 'to' ID %s: %v", commandsComponents[2], err)
			}

			if _, ok := idToCommandIndex[CommandID(to)]; !ok {
				commands = append(commands, &Command{
					ID:      CommandID(to),
					Bucket:  make([]*Command, 0),
					Parents: make([]*Command, 0),
					Nexts:   make([]*Command, 0),
				})

				idToCommandIndex[CommandID(to)] = len(commands) - 1
			}

			if commandsComponents[1] == "JUMP" {
				prevCommand = nil
			} else {
				prevCommand = newCommand
			}

			next := commands[idToCommandIndex[CommandID(to)]]
			newCommand.Nexts = append(newCommand.Nexts, next)
			next.Parents = append(next.Parents, newCommand)
		default:
			log.Fatalf("unknown command %s", commandsComponents[1])
		}
	}

	dfs(firstCommand)
	sort.Sort(CommandSlice(commands))
	dominators(firstCommand, commands)

	var naturalCycles int
	for _, current := range commands {
		wasFirst := false
		if current.TimeIn == 0 {
			break
		}

	PARENTS_CYCLE:
		for _, parent := range current.Parents {
			if parent.TimeIn == 0 {
				continue
			}

			//Поднимаеся по дереву доминаторов
			for idom := parent; idom != nil; idom = idom.Idom {
				if idom == current {
					wasFirst = true
					break PARENTS_CYCLE
				}
			}
		}

		if wasFirst {
			naturalCycles++
		}
	}

	fmt.Println(naturalCycles)
}
