package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
)

var numServers int
var numWebServers int

func init() {
	flag.IntVar(&numServers, "numServers", 2, "The number of dlog servers to use")
	flag.IntVar(&numWebServers, "numWebServers", 2, "The number of web servers to use")
}

func main() {
	flag.PrintDefaults()
	flag.Parse()

	cmds := "docker build -t dlog-exp ."

	servers := make([]string, numServers)
	serverLinks := make([]string, numServers)
	serverPorts := make([]string, numServers)
	for i := 0; i < numServers; i++ {
		servers[i] = "s" + strconv.Itoa(i)
		serverLinks[i] = servers[i] + ":" + servers[i]
		serverPorts[i] = servers[i] + ":1234"
		cmds += fmt.Sprintf(";docker run -d --name %s --expose 1234 -P dlog-exp server", servers[i])
	}

	webServers := make([]string, numWebServers)
	webServerLinks := make([]string, numWebServers)
	for i := 0; i < numWebServers; i++ {
		webServers[i] = "ws_" + strconv.Itoa(i)
		webServerLinks[i] = webServers[i] + ":" + webServers[i]
		cmds += fmt.Sprintf(";docker run -d --name %s --expose 80 %s dlog-exp webserver -servers %s", webServers[i], "--link="+strings.Join(serverLinks, " --link="), strings.Join(serverPorts, ","))
	}

	cmds += fmt.Sprintf(";docker run -d --name lb -e WS_PATH=/ %s -P jasonwyatt/nginx-loadbalancer", "--link="+strings.Join(webServerLinks, " --link="))

	for _, name := range append(append(servers, webServers...), "lb") {
		runCmd(exec.Command("docker", "stop", name))
		runCmd(exec.Command("docker", "rm", name))
	}

	//Run commands
	for _, cmdLine := range strings.Split(cmds, ";") {
		cmd := strings.Split(cmdLine, " ")
		runCmd(exec.Command(cmd[0], cmd[1:]...))
	}
	//Print ports
	for i := 0; i < numServers; i++ {
		runCmd(exec.Command("docker", "port", servers[i]))
	}
	runCmd(exec.Command("docker", "port", "lb"))

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	s := <-c
	fmt.Println(s)
}

func runCmd(cmd *exec.Cmd) {
	fmt.Println(strings.Join(cmd.Args, " "))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}

}
