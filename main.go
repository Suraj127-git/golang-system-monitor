package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

type model struct {
	cpuUsage  float64
	ramUsage  float64
	diskUsage float64
	tasks     []string
	quitting  bool
	err       error
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
	)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return t
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case time.Time:
		// Fetch CPU usage
		cpuPercents, err := cpu.Percent(0, false)
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		m.cpuUsage = cpuPercents[0]

		// Fetch RAM usage
		vMem, err := mem.VirtualMemory()
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		m.ramUsage = vMem.UsedPercent

		// Fetch Disk usage
		diskUsage, err := disk.Usage("/")
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		m.diskUsage = diskUsage.UsedPercent

		// Fetch running tasks
		processes, err := process.Processes()
		if err != nil {
			m.err = err
			return m, tea.Quit
		}
		m.tasks = []string{}
		for _, p := range processes {
			name, err := p.Name()
			if err == nil {
				m.tasks = append(m.tasks, name)
			}
		}

		return m, tick()
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	if m.quitting {
		return "Quitting...\n"
	}

	var tasks string
	for _, task := range m.tasks {
		tasks += fmt.Sprintf("%s\n", task)
	}

	return fmt.Sprintf(
		"CPU Usage: %.2f%%\nRAM Usage: %.2f%%\nDisk Usage: %.2f%%\n\nRunning Tasks:\n%s\nPress q to quit.",
		m.cpuUsage, m.ramUsage, m.diskUsage, tasks,
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
