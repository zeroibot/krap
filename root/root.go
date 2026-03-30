package root

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/zeroibot/fn/dict"
	"github.com/zeroibot/fn/ds"
	"github.com/zeroibot/fn/fail"
	"github.com/zeroibot/fn/lang"
	"github.com/zeroibot/fn/str"
	"golang.org/x/term"
)

const (
	allCommands string = "*"
	cmdHelp     string = "help"
	cmdExit     string = "exit"
	cmdSearch   string = "cmd"
	cmdGlue     string = "/"
)

var cmdMap = map[string]*CmdConfig{}

var (
	errInvalidCommand    = errors.New("invalid command")
	errInvalidParamCount = errors.New("invalid param count")
	getHelp              = fmt.Sprintf("Type `%s` for list of commands, `%s <keyword>` to search for command", cmdHelp, cmdSearch)
	helpSkipCommands     = []string{cmdHelp, cmdExit, cmdSearch}
)

// Takes in list of parameters
type CmdHandler = func([]string)

type CmdConfig struct {
	Command   string
	MinParams int
	Docs      string
	Handler   CmdHandler
}

// Creates a new CmdConfig
func NewCommand(command string, minParams int, docs string, handler CmdHandler) *CmdConfig {
	return &CmdConfig{command, minParams, docs, handler}
}

// Creates a new map of command => CmdConfigs
func NewCommandMap(cfgs ...*CmdConfig) map[string]*CmdConfig {
	commands := make(map[string]*CmdConfig)
	for _, cfg := range cfgs {
		commands[cfg.Command] = cfg
	}
	return commands
}

// Sets the command map
func SetCommandMap(commands map[string]*CmdConfig) {
	cmdMap = commands
}

// Root's main loop
func MainLoop(onExit func()) {
	var err error
	var line, command string
	var params []string

	fmt.Println("Commands:", len(cmdMap))
	fmt.Printf("Root: type `%s` for list of commands, `%s` to close\n", cmdHelp, cmdExit)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n> ")
		line, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		command, params = getCommandParams(line)
		if command == "" {
			continue
		}
		switch command {
		case cmdExit:
			if onExit != nil {
				onExit()
			}
			return
		case cmdHelp:
			if len(params) == 0 {
				command = allCommands
			} else {
				command = params[0]
			}
			displayHelp(command)
		case cmdSearch:
			keyword := params[0]
			searchCommand(keyword)
		default:
			c := cmdMap[command]
			c.Handler(params)
		}
	}
}

// Get key=value map from parameters list
func ParamsMap(params []string, required []string, optional []string) (dict.StringMap, error) {
	if required == nil {
		required = make([]string, 0)
	}
	if optional == nil {
		optional = make([]string, 0)
	}
	paramsMap := make(dict.StringMap)
	for _, param := range params {
		parts := str.CleanSplitN(param, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		if !slices.Contains(required, key) && !slices.Contains(optional, key) {
			continue
		}
		paramsMap[key] = value
	}
	for _, key := range required {
		if _, ok := paramsMap[key]; !ok {
			return nil, fail.MissingParams
		}
	}
	return paramsMap, nil
}

// Authenticate Root account in command-line app
func Authenticate(authFn func(string) error) {
	fmt.Print("Enter password: ")
	pwd, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Println()
	password := strings.TrimSpace(string(pwd))
	err = authFn(password)
	if err != nil {
		log.Fatal("Root authentication failed")
	}
}

// Check if the command exists and if it meets the min parameter count
func validateCommandParams(command string, params []string) error {
	if command == cmdExit || command == cmdHelp {
		return nil
	}
	if command == cmdSearch {
		return lang.Ternary(len(params) < 1, errInvalidParamCount, nil)
	}
	cfg, ok := cmdMap[command]
	if !ok {
		return errInvalidCommand
	}
	if len(params) < cfg.MinParams {
		return errInvalidParamCount
	}
	return nil
}

// Get command and params from line
func getCommandParams(line string) (string, []string) {
	if strings.TrimSpace(line) == "" {
		fmt.Println(getHelp)
		return "", nil
	}
	args := str.SpaceSplit(line)
	command, params := args[0], args[1:]
	command = strings.ToLower(command)
	err := validateCommandParams(command, params)
	if err != nil {
		fmt.Println("Error:", err)
		if errors.Is(err, errInvalidCommand) {
			fmt.Println(getHelp)
		} else if errors.Is(err, errInvalidParamCount) {
			displayHelp(command)
		}
		return "", nil
	}
	return command, params
}

// Display help list
func displayHelp(targetCommand string) {
	targetCommand = strings.ToLower(targetCommand)
	if _, ok := cmdMap[targetCommand]; !ok && targetCommand != allCommands && !slices.Contains(helpSkipCommands, targetCommand) {
		fmt.Println("Error: unknown command: ", targetCommand)
		fmt.Println(getHelp)
		return
	}
	fmt.Println("Usage: <command> <params>")
	fmt.Println("\nCommands and params:")

	commands := dict.Keys(cmdMap)
	sort.Strings(commands)
	for _, command := range commands {
		if slices.Contains(helpSkipCommands, command) {
			continue
		}
		cfg := cmdMap[command]
		if targetCommand == allCommands || targetCommand == command {
			fmt.Printf("%-30s\t%s\n", command, cfg.Docs)
		}
	}
}

// Search for command keyword
func searchCommand(keyword string) {
	keyword = strings.ToLower(keyword)
	commands := dict.Keys(cmdMap)
	slices.Sort(commands)
	if keyword == allCommands {
		stems := ds.NewSet[string]()
		for _, command := range commands {
			if slices.Contains(helpSkipCommands, command) {
				continue
			}
			stem := str.CleanSplit(command, cmdGlue)[0]
			stems.Add(stem)
		}
		heads := stems.Items()
		slices.Sort(heads)
		for _, head := range heads {
			fmt.Println(head)
		}
	} else {
		for _, command := range commands {
			if slices.Contains(helpSkipCommands, command) {
				continue
			}
			if strings.Contains(command, keyword) {
				cfg := cmdMap[command]
				fmt.Printf("%-30s\t%s\n", command, cfg.Docs)
			}
		}
	}
}
