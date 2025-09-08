package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	envInfo := collectEnvInfo()
	err := checkEssentialCommands(envInfo)
	if err != nil {
		fmt.Println(err)
		return
	}
	initConfig := PromptInitConfig(envInfo)
	initProject(initConfig)
}

func initProject(initConfig InitConfig) {

	fmt.Println("Initializing go module...")
	cmd := exec.Command("go", "mod", "init", initConfig.GoModName)
	if cmd.Err != nil {
		fmt.Println("Error initializing go module")
		fmt.Println(cmd.Err)
		return
	}

	fmt.Println("Done")
	fmt.Println("Initializing directory structure...")
	os.MkdirAll("./cmd/"+filepath.Base(initConfig.GoModName), os.ModePerm)
	os.MkdirAll("./pkg", os.ModePerm)
	os.MkdirAll("./internal", os.ModePerm)
	fmt.Println("Done")

	fmt.Println("Creating main.go...")
	mainFile, err := os.Create("./cmd/" + filepath.Base(initConfig.GoModName) + "/main.go")
	if err != nil {
		fmt.Println("Error creating main.go")
		fmt.Println(err)
		return
	}
	defer mainFile.Close()
	mainFile.WriteString("package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n\tos.Exit(0)\n}\n")
	mainFile.Sync()
	mainFile.Close()
	fmt.Println("Done")

	if initConfig.InitTask {
		fmt.Println("Initializing taskfile...")
		if err := exec.Command("task", "--init").Run(); err != nil {
			fmt.Println("Error initializing taskfile")
			fmt.Println(err)
		} else {
			fmt.Println("Done")
		}
	}

	if initConfig.InitMake {
		fmt.Println("Initializing makefile...")
		os.WriteFile("./Makefile", []byte("all:\n\tgo run ./cmd/"+filepath.Base(initConfig.GoModName)+"/main.go\n"), os.ModePerm)
		fmt.Println("Done")
	}

	if initConfig.ReInitGit {
		fmt.Println("Reinitializing git...")
		if err := exec.Command("git", "init").Run(); err != nil {
			fmt.Println("Error reinitializing git")
			fmt.Println(err)
		} else {
			fmt.Println("Done")
		}
		if initConfig.GitProvider != "" {
			fmt.Println("Configuring git...")
			fmt.Println("git remote add origin " + initConfig.GitPath)
			if err := exec.Command("git", "remote", "add", "origin", initConfig.GitPath).Run(); err != nil {
				fmt.Println("Error configuring git")
				fmt.Println(err)
			} else {
				fmt.Println("Done")
			}
		}
	}

	if initConfig.RemoveInit {
		fmt.Println("Removing init.go...")
		os.Remove("./init.go")
		fmt.Println("Done")
	}

	fmt.Println("Initialization finished!")
	fmt.Println("You can now run 'go run ./cmd/" + filepath.Base(initConfig.GoModName) + "/main.go' to run your program")
	fmt.Println("Or you can run 'go build ./cmd/" + filepath.Base(initConfig.GoModName) + "/main.go' to build your program")
}

type InitConfig struct {
	GitProvider string
	GitUsername string
	GitPath     string
	GoModName   string
	InitTask    bool
	InitMake    bool
	ReInitGit   bool
	RemoveInit  bool
}

func YesNoPrompt(question string, defaultYes bool) bool {
	if defaultYes {
		question += " (default: y)"
	} else {
		question += " (default: n)"
	}
	fmt.Println(question + " (y/n)")
	var answer string
	fmt.Scanln(&answer)
	if answer == "n" {
		return false
	}
	if answer == "y" {
		return true
	}
	return defaultYes
}

func PromptInitConfig(envInfo EnvInfo) InitConfig {
	var initConfig InitConfig
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working directory")
		fmt.Println(err)
	} else {
		initConfig.GoModName = filepath.Base(wd)
	}
	if envInfo.GitInstalled {
		bytes, err := exec.Command("git", "config", "user.name").Output()
		if err != nil {
			bytes = []byte{}
		}
		maybeUsername := strings.TrimSpace(string(bytes))
		fmt.Println("What's your git username? (default: " + maybeUsername + ")")
		var gitUsername string
		fmt.Scanln(&gitUsername)
		if gitUsername == "" {
			gitUsername = maybeUsername
		}
		initConfig.GitUsername = gitUsername

		gitProvider := ""
		fmt.Println("What's your git provider? (default: none)")
		var gitProviderInput string
		fmt.Scanln(&gitProviderInput)
		if gitProviderInput != "" {
			gitProvider = gitProviderInput
		}
		initConfig.GitProvider = gitProvider
		if gitProvider != "" {
			initConfig.GitPath = "https://" + gitProvider + "/" + gitUsername + "/"
		}
	}
	if initConfig.GitProvider != "" {
		curdir, _ := os.Getwd()
		curdir = filepath.Base(curdir)
		fmt.Println("Your go module name? (default: " + initConfig.GitPath + curdir + ")")
		var goModuleName string
		fmt.Scanln(&goModuleName)
		if goModuleName != "" {
			initConfig.GoModName = goModuleName
		} else {
			initConfig.GoModName = initConfig.GitPath + curdir
		}
	} else {
		fmt.Println("Your go module name? (default: " + initConfig.GoModName + ")")
		var goModuleName string
		fmt.Scanln(&goModuleName)
		if goModuleName != "" {
			initConfig.GoModName = goModuleName
		}
	}

	initConfig.InitTask = envInfo.TaskInstalled && YesNoPrompt("Do you want to initialize a taskfile? (y/n)", true)
	initConfig.InitMake = envInfo.MakeInstalled && YesNoPrompt("Do you want to initialize a makefile? (y/n)", true)
	initConfig.ReInitGit = envInfo.GitInstalled && YesNoPrompt("Do you want to reinitialize git? (y/n)", true)
	initConfig.RemoveInit = YesNoPrompt("Do you want to remove init.go? (y/n)", true)

	return initConfig
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

type EnvInfo struct {
	TaskInstalled bool
	GoInstalled   bool
	MakeInstalled bool
	GitInstalled  bool
}

func collectEnvInfo() EnvInfo {
	var envInfo EnvInfo
	if commandExists("task") {
		envInfo.TaskInstalled = true
	}
	if commandExists("go") {
		envInfo.GoInstalled = true
	}
	if commandExists("make") {
		envInfo.MakeInstalled = true
	}
	if commandExists("git") {
		envInfo.GitInstalled = true
	}
	return envInfo
}

func checkEssentialCommands(envInfo EnvInfo) error {
	if !envInfo.GoInstalled {
		return errors.New("go is not installed")
	}
	return nil
}
