package main

import (
	"errors"
	"fmt"
	"log"
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
	fmt.Println("go mod init " + initConfig.GoModName)
	_, err := exec.Command("go", "mod", "init", initConfig.GoModName).Output()
	if err != nil {
		log.Println("Error initializing go module")
		log.Println(err)
		return
	}

	fmt.Println("Done")

	if initConfig.DirectoryName != "gowebtemplate" {
		fmt.Println("renaming directory...")
		if err := os.Rename(".", initConfig.DirectoryName); err != nil {
			log.Println("Error renaming current directory")
			log.Println(err)
		}
	}
	fmt.Println("Initializing directory structure...")
	os.MkdirAll("./cmd/"+filepath.Base(initConfig.GoModName), os.ModePerm)
	os.MkdirAll("./pkg", os.ModePerm)
	os.MkdirAll("./internal", os.ModePerm)
	fmt.Println("Done")

	fmt.Println("Creating main.go...")
	mainFile, err := os.Create("./cmd/" + filepath.Base(initConfig.GoModName) + "/main.go")
	if err != nil {
		log.Println("Error creating main.go")
		log.Println(err)
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
			log.Println("Error initializing taskfile")
			log.Println(err)
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
			log.Println("Error reinitializing git")
			log.Println(err)
		} else {
			fmt.Println("Done")
		}
		if initConfig.GitRepo != "" {
			fmt.Println("Configuring git...")
			fmt.Println("git remote set-url origin " + initConfig.GitRepo)
			if err := exec.Command("git", "remote", "set-url", "origin", initConfig.GitRepo).Run(); err != nil {
				log.Println("Error configuring git")
				log.Println(err)
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
	GitRepo       string
	GoModName     string
	InitTask      bool
	InitMake      bool
	ReInitGit     bool
	RemoveInit    bool
	DirectoryName string
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
	wd = filepath.Base(wd)
	if envInfo.GitInstalled {
		fmt.Println("What's your git repository?")
		var gitRepo string
		fmt.Scanln(&gitRepo)
		if gitRepo != "" {
			initConfig.GitRepo = gitRepo
		}
	}
	if initConfig.GitRepo != "" {
		gitModPath := strings.TrimPrefix(initConfig.GitRepo, "https://")
		gitModPath = strings.TrimPrefix(gitModPath, "http://")
		initConfig.GoModName = gitModPath
	}

	fmt.Println("Your go module name? (default: " + initConfig.GoModName + ")")
	var goModuleName string
	fmt.Scanln(&goModuleName)
	if goModuleName != "" {
		initConfig.GoModName = goModuleName
	}

	if initConfig.GitRepo == "" && YesNoPrompt("Do you want to rename directory to "+initConfig.GoModName+" (y/n) (default: n)", false) {
		initConfig.DirectoryName = initConfig.GoModName
	} else {
		initConfig.DirectoryName = "gowebtemplate"
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
