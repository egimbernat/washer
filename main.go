package main

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	linkPath     string
	hostFilePath string
)

type Links struct {
	Link []struct {
		ActorID   string `toml:"actor-id"`
		ActorName string `toml:"actor-name"`
		Provider  string `toml:"provider"`
		Link      string `toml:"link"`
		Contract  string `toml:"contract"`
		Values    string `toml:"values"`
	} `toml:"link"`
}

type Hosts struct {
	Success bool `json:"success"`
	Hosts   []struct {
		ID            string `json:"id"`
		UptimeSeconds int    `json:"uptime_seconds"`
	} `json:"hosts"`
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "washer",
		Short: "Extension for wash",
		Long:  `Extends wash for production operations`,
		Run: func(cmd *cobra.Command, args []string) {
			for _, host := range getHosts().Hosts {
				res, err := exec.Command("wash", "ctl", "stop", "host", host.ID, "--host-timeout", "10000").Output()
				if err != nil {
					fmt.Println(string(res))
					log.Fatal(err.Error())
					os.Exit(1)
				}
				fmt.Println(string(res))
				fmt.Println("Waiting to new host being online...")

				successful := false
				for i := 0; i < 60; i++ {
					if successful {
						break
					}
					time.Sleep(10 * time.Second)
					for _, h := range getHosts().Hosts {
						if h.UptimeSeconds > 30 && h.UptimeSeconds < 120 {
							//Is pretty sure this is new host
							res, err := exec.Command("wash", "ctl", "apply", h.ID, hostFilePath).Output()
							if err != nil {
								fmt.Println(string(res))
								log.Fatal(err.Error())
								os.Exit(1)
							}
							fmt.Println(string(res))
							successful = true
							break
						}
					}
				}
				if !successful {
					log.Fatal("Host doesn't return online")
				}
			}
		},
	}

	rootCmd.Flags().StringVarP(&hostFilePath, "path", "p", "", "")

	linkCmd := link()
	linkCmd.Flags().StringVarP(&linkPath, "path", "p", "", "")
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(reLink())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getHosts() Hosts {
	output, err := exec.Command("wash", "ctl", "get", "hosts", "-o", "json").Output()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	var hosts Hosts
	err = json.Unmarshal(output, &hosts)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	fmt.Println(hosts)
	return hosts
}

func link() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Link all actors",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			file, err := ioutil.ReadFile(linkPath)
			if err != nil {
				return
			}

			var links Links

			_, err = toml.Decode(string(file), &links)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(links)

			for _, l := range links.Link {
				args := []string{"ctl", "link", "put", "--link-name", l.Link, l.ActorID, l.Provider, l.Contract}
				if len(l.Values) > 0 {
					args = append(args, strings.Split(l.Values, ",")...)
				}
				res, err := exec.Command("wash", args...).Output()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Println(string(res))
			}
		},
	}
}

func reLink() *cobra.Command {
	type Links struct {
		Links []struct {
			ActorID    string `json:"actor_id"`
			ContractID string `json:"contract_id"`
			LinkName   string `json:"link_name"`
			ProviderID string `json:"provider_id"`
		} `json:"links"`
		Success bool `json:"success"`
	}
	return &cobra.Command{
		Use:   "unlink",
		Short: "Unlike all actors",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			output, err := exec.Command("wash", "ctl", "link", "query", "-o", "json").Output()
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			var links Links
			err = json.Unmarshal(output, &links)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			for _, l := range links.Links {
				res, err := exec.Command("wash", "ctl", "link", "del", "-l", l.LinkName, l.ActorID, l.ContractID).Output()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Println(string(res))
			}
		},
	}
}
