package main

import (
	"fmt"
	//"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dockercn/docker-bucket/drone/pkg/build"
	"github.com/dockercn/docker-bucket/drone/pkg/build/docker"
	"github.com/dockercn/docker-bucket/drone/pkg/build/log"
	"github.com/dockercn/docker-bucket/drone/pkg/build/repo"
	"github.com/dockercn/docker-bucket/drone/pkg/build/script"
)

var (
	// version number, currently deterined by the
	// git revision number (sha)
	version string
)

func init() {
	// default logging
	log.SetPrefix("\033[2m[DRONE] ")
	log.SetSuffix("\033[0m\n")
	log.SetOutput(os.Stdout)
	log.SetPriority(log.LOG_NOTICE)
}

func drone(yaml string) {
	s, err := script.ParseBuildFile(yaml)
	if err != nil {
		log.Err(err.Error())
		os.Exit(1)
		return
	}
	fmt.Println(s.Dependencies)
	for _, f := range s.Dependencies {
		run("../tests/drone/" + f)
	}
	run(yaml)
}

func main() {
	yaml := "../tests/drone/sample.yml"
	drone(yaml)
}

func run(path string) {
	dockerClient := docker.New()

	// parse the Drone yml file
	s, err := script.ParseBuildFile(path)
	//fmt.Println(s.Repo)
	if err != nil {
		log.Err(err.Error())
		os.Exit(1)
		return
	}

	// get the repository url
	// here we should use githubapi to accplish this
	//dir := filepath.Dir(path)
	code := repo.Repo{
		Name:   "test",
		Branch: "HEAD",
		Path:   s.Repo,
	}

	// this is where the code gets uploaded to the container
	// TODO move this code to the build package
	code.Dir = filepath.Join("/var/cache/drone/src", s.Repo[5:])

	// track all build results
	var builders []*build.Builder

	//here we should parse all the builds, first the deps.
	builds := []*script.Build{s}

	// loop through and create builders
	for _, b := range builds { //script.Builds {
		builder := build.New(dockerClient)
		builder.Build = b
		builder.Repo = &code
		//builder.Key = key
		builder.Stdout = os.Stdout
		builder.Timeout = 300 * time.Minute
		//builder.Privileged = *privileged

		//if *parallel == true {
		//var buf bytes.Buffer
		//builder.Stdout = &buf
		//}

		builders = append(builders, builder)
	}

	//switch *parallel {
	//case false:
	runSequential(builders)
	//case true:
	//runParallel(builders)
	//}

	// this exit code is initially 0 and will
	// be set to an error code if any of the
	// builds fail.
	var exit int

	fmt.Printf("\nDrone Build Results \033[90m(%v)\033[0m\n", len(builders))

	// loop through and print results
	for _, builder := range builders {
		build := builder.Build
		res := builder.BuildState
		duration := time.Duration(res.Finished - res.Started)
		switch {
		case builder.BuildState.ExitCode == 0:
			fmt.Printf(" \033[32m\u2713\033[0m %v \033[90m(%v)\033[0m\n", build.Name, humanizeDuration(duration*time.Second))
		case builder.BuildState.ExitCode != 0:
			fmt.Printf(" \033[31m\u2717\033[0m %v \033[90m(%v)\033[0m\n", build.Name, humanizeDuration(duration*time.Second))
			exit = builder.BuildState.ExitCode
		}
	}

	os.Exit(exit)
}

func runSequential(builders []*build.Builder) {
	// loop through and execute each build
	for _, builder := range builders {
		if err := builder.Run(); err != nil {
			log.Errf("Error executing build: %s", err.Error())
			os.Exit(1)
		}
	}
}

func runParallel(builders []*build.Builder) {
	// spawn four worker goroutines
	var wg sync.WaitGroup
	for _, builder := range builders {
		// Increment the WaitGroup counter
		wg.Add(1)
		// Launch a goroutine to run the build
		go func(builder *build.Builder) {
			defer wg.Done()
			builder.Run()
		}(builder)
		time.Sleep(500 * time.Millisecond) // get weird iptables failures unless we sleep.
	}

	// wait for the workers to finish
	wg.Wait()
}
