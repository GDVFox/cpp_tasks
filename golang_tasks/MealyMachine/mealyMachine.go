package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/skorobogatov/input"
)

//Лес непересекающихся множеств.
type dsu struct {
	parents []int
	depth   []int
}

func initDSU(size int) *dsu {
	return &dsu{
		parents: make([]int, size),
		depth:   make([]int, size),
	}
}

func (d *dsu) MakeSet(x int) {
	d.parents[x] = x
	d.depth[x] = 0
}

func (d *dsu) Find(x int) int {
	if d.parents[x] == x {
		return x
	}
	d.parents[x] = d.Find(d.parents[x])
	return d.parents[x]
}

func (d *dsu) Unite(x, y int) {
	x = d.Find(x)
	y = d.Find(y)

	if d.depth[x] < d.depth[y] {
		d.parents[x] = y
	} else {
		d.parents[y] = x
		if d.depth[x] == d.depth[y] {
			d.depth[x]++
		}
	}
}

//MealyMachine представление автомата Мили,
//используем для dependency injection.
type MealyMachine struct {
	StateNum         int
	AlphabetSize     int
	Start            int
	TransitionMatrix [][]int
	ExitMatrix       [][]string

	pos        []int
	currentPos int
}

func (m *MealyMachine) dfs(currentState int) {
	m.pos[currentState] = m.currentPos
	m.currentPos++
	for _, target := range m.TransitionMatrix[currentState] {
		if m.pos[target] == -1 {
			m.dfs(target)
		}
	}
}

//CaninicalEnumirate строит каноническую нумерацию состояний автомата
func (m *MealyMachine) CaninicalEnumirate() {
	m.currentPos = 0
	m.pos = make([]int, m.StateNum)
	for i := range m.pos {
		m.pos[i] = -1
	}

	m.dfs(m.Start)
	newTransitionMatrix := make([][]int, m.currentPos)
	newExitMatrix := make([][]string, m.currentPos)
	for i := range newTransitionMatrix {
		newTransitionMatrix[i] = make([]int, m.AlphabetSize)
		newExitMatrix[i] = make([]string, m.AlphabetSize)
	}

	for i := 0; i < m.StateNum; i++ {
		if m.pos[i] == -1 {
			continue
		}
		for j := 0; j < m.AlphabetSize; j++ {
			newTransitionMatrix[m.pos[i]][j] = m.pos[m.TransitionMatrix[i][j]]
			newExitMatrix[m.pos[i]][j] = m.ExitMatrix[i][j]
		}
	}

	m.Start = 0
	m.StateNum = m.currentPos
	m.TransitionMatrix = newTransitionMatrix
	m.ExitMatrix = newExitMatrix
}

func (m *MealyMachine) String() string {
	var b bytes.Buffer
	b.WriteString(strconv.Itoa(m.StateNum) + "\n")
	b.WriteString(strconv.Itoa(m.AlphabetSize) + "\n")
	b.WriteString(strconv.Itoa(m.Start) + "\n")
	for _, row := range m.TransitionMatrix {
		for _, elem := range row {
			b.WriteString(strconv.Itoa(elem) + " ")
		}
		b.WriteByte('\n')
	}
	for _, row := range m.ExitMatrix {
		b.WriteString(strings.Join(row, " "))
		b.WriteByte('\n')
	}

	return b.String()
}

//ToDOT преобразовывает автомат в DOT формат для обработки graphviz
func (m *MealyMachine) ToDOT() string {
	var b bytes.Buffer
	b.WriteString("digraph {\n")
	b.WriteString("\trankdir = LR\n")
	b.WriteString("\t" + `dummy [label = "", shape = none]` + "\n")
	for i := 0; i < m.StateNum; i++ {
		b.WriteByte('\t')
		b.WriteString(strconv.Itoa(i))
		b.WriteString(" [shape = circle]\n")
	}
	b.WriteString("\tdummy -> ")
	b.WriteString(strconv.Itoa(m.Start))
	b.WriteByte('\n')
	for i := 0; i < m.StateNum; i++ {
		for j := 0; j < m.AlphabetSize; j++ {
			b.WriteByte('\t')
			b.WriteString(strconv.Itoa(i))
			b.WriteString(" -> ")
			b.WriteString(strconv.Itoa(m.TransitionMatrix[i][j]))
			b.WriteString(` [label = "`)
			b.WriteByte('a' + byte(j))
			b.WriteString("(" + m.ExitMatrix[i][j] + ")")
			b.WriteString(`"]`)
			b.WriteByte('\n')
		}
	}
	b.WriteString("}")

	return b.String()
}

func (m *MealyMachine) split1(d *dsu, roots []int) int {
	newStateNum := m.StateNum
	for i := 0; i < m.StateNum; i++ {
		d.MakeSet(i)
	}

	for i := 0; i < m.StateNum; i++ {
		for j := 0; j < m.StateNum; j++ {
			if d.Find(i) != d.Find(j) {
				eqFlag := true
				for k := 0; k < m.AlphabetSize; k++ {
					if m.ExitMatrix[i][k] != m.ExitMatrix[j][k] {
						eqFlag = false
						break
					}
				}
				if eqFlag {
					d.Unite(i, j)
					newStateNum--
				}
			}
		}
	}

	for i := 0; i < m.StateNum; i++ {
		roots[i] = d.Find(i)
	}

	return newStateNum
}

func (m *MealyMachine) split(d *dsu, roots []int) int {
	newStateNum := m.StateNum
	for i := 0; i < m.StateNum; i++ {
		d.MakeSet(i)
	}

	for i := 0; i < m.StateNum; i++ {
		for j := 0; j < m.StateNum; j++ {
			if roots[i] == roots[j] && d.Find(i) != d.Find(j) {
				eqFlag := true
				for k := 0; k < m.AlphabetSize; k++ {
					tr1, tr2 := m.TransitionMatrix[i][k], m.TransitionMatrix[j][k]
					if roots[tr1] != roots[tr2] {
						eqFlag = false
						break
					}
				}
				if eqFlag {
					d.Unite(i, j)
					newStateNum--
				}
			}
		}
	}

	for i := 0; i < m.StateNum; i++ {
		roots[i] = d.Find(i)
	}

	return newStateNum
}

//Minimize выполняет минимизацию автомата Мили,
//используя алгоритм Ауфенкампа-Хона
func (m *MealyMachine) Minimize() {
	d := initDSU(m.StateNum)
	roots := make([]int, m.StateNum)

	m.split1(d, roots)
	oldNewStateNum := -1
	for {
		newStateNum := m.split(d, roots)
		if newStateNum == oldNewStateNum {
			break
		}
		oldNewStateNum = newStateNum
	}

	//Сжимаем roots, чтобы попадать в новые размеры.
	rootSave := make(map[int]int)
	var counter int
	for i := 0; i < m.StateNum; i++ {
		if savedRoot, ok := rootSave[roots[i]]; !ok {
			rootSave[roots[i]] = counter
			roots[i] = counter
			counter++
		} else {
			roots[i] = savedRoot
		}
	}

	newTransitionMatrix := make([][]int, 0, oldNewStateNum)
	newExitMatrix := make([][]string, 0, oldNewStateNum)
	for i := 0; i < m.StateNum; i++ {
		root := roots[i]
		if len(newTransitionMatrix) == root {
			newTransitionMatrix = append(newTransitionMatrix, make([]int, m.AlphabetSize))
			newExitMatrix = append(newExitMatrix, make([]string, m.AlphabetSize))
			for j := 0; j < m.AlphabetSize; j++ {
				newTransitionMatrix[len(newTransitionMatrix)-1][j] = roots[m.TransitionMatrix[i][j]]
				newExitMatrix[len(newExitMatrix)-1][j] = m.ExitMatrix[i][j]
			}
		}
	}

	m.Start = roots[m.Start]
	m.StateNum = oldNewStateNum
	m.TransitionMatrix = newTransitionMatrix
	m.ExitMatrix = newExitMatrix
}

func readMealyMachine(machine *MealyMachine) {
	//Игнорирую все ошибки при чтении
	input.Scanf("%d %d %d", &machine.StateNum, &machine.AlphabetSize, &machine.Start)
	machine.TransitionMatrix = make([][]int, machine.StateNum)
	machine.ExitMatrix = make([][]string, machine.StateNum)
	for i := range machine.TransitionMatrix {
		machine.TransitionMatrix[i] = make([]int, machine.AlphabetSize)
		machine.ExitMatrix[i] = make([]string, machine.AlphabetSize)
	}
	for i := 0; i < machine.StateNum; i++ {
		for j := 0; j < machine.AlphabetSize; j++ {
			input.Scanf("%d", &machine.TransitionMatrix[i][j])
		}
	}
	for i := 0; i < machine.StateNum; i++ {
		for j := 0; j < machine.AlphabetSize; j++ {
			input.Scanf("%s", &machine.ExitMatrix[i][j])
		}
	}

}

func main() {
	//Пример использования затрагивающий практически все методы автомата
	//Сравниение двух автоматов Мили на эквивалентность.
	machine1 := &MealyMachine{}
	machine2 := &MealyMachine{}

	readMealyMachine(machine1)
	readMealyMachine(machine2)

	machine1.Minimize()
	machine1.CaninicalEnumirate()
	machine2.Minimize()
	machine2.CaninicalEnumirate()

	if machine1.ToDOT() == machine2.ToDOT() {
		fmt.Println("EQUAL")
	} else {
		fmt.Println("NOT EQUAL")
	}
}
