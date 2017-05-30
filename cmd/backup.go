package cmd

import (
	"github.com/spf13/cobra"
	"fmt"
	"errors"
	"github.com/levigross/grequests"
	"github.com/PuerkitoBio/goquery"
	"log"
	"time"
	"os"
)

var (
	hostname,
	username,
	password string
)


// Get the csrf token for requests
func CSRF(r *grequests.Response) string {
	doc, _ := goquery.NewDocumentFromResponse(r.RawResponse)
	csrfToken, _ := doc.Find("input[name=__csrf_magic]").Attr("value")
	return csrfToken
}


func backup(h, u, p string) (bool, error) {
	session := grequests.NewSession(nil)
	visit, err := session.Get("https://" + h + "/index.php", nil)
	if err != nil {
		return false, errors.New("Unable to reach host!")
	}
	defer visit.Close()
	csrfToken := CSRF(visit)
	loginPayload := &grequests.RequestOptions{Data: map[string]string{
		"__csrf_magic": csrfToken,
		"usernamefld": u,
		"passwordfld": p,
		"login": "",
	}}
	r, err := session.Post("https://" + h + "/index.php", loginPayload)
	if err != nil {
		log.Fatal("Unable to login!")
		return false, errors.New("Unable to login!")
	}
	defer r.Close()
	backupPage, _ := session.Get("https://" + h + "/diag_backup.php", nil)
	defer backupPage.Close()
	csrfToken = CSRF(backupPage)
	backupPayload := &grequests.RequestOptions{Data: map[string]string{
		"__csrf_magic": csrfToken,
		"backuparea": "",
		"donotbackuprrd": "yes",
		"download": "Download configuration as XML",
	}}
	backupFile, _ := session.Post("https://" + h + "/diag_backup.php", backupPayload)
	defer backupFile.Close()
	current_time := time.Now().Local()
	today := current_time.Format("2006_01_02")
	backupFile.DownloadToFile(today + "_" + h + ".xml")
	log.Println("Backup saved as: " + today + "_" + h + ".xml")
	return true, nil
}



// serveCmd represents the serve command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup pfSense config as XML",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(hostname) == 0 {
			//err := errors.New("Hostname cannot be empty!")
			//fmt.Printf("%v\n", err)
			RootCmd.Help()
			os.Exit(1)
		}
		fmt.Printf("Backup host: %s\n", hostname)
		backup(hostname, username, password)
	},
}


func init() {
	RootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&hostname, "hostname", "H", "", "Hostname of pfSense target.")
	backupCmd.Flags().StringVarP(&username, "username", "u", "", "Username for authentication.")
	backupCmd.Flags().StringVarP(&password, "password", "p", "", "Password for authentication.")

	//laber.AddCommand(cmdSay)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

