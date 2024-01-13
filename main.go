package main

import (
	"fmt"
	"github.com/gtuk/discordwebhook"
	"github.com/tidwall/gjson"
	"io"
	"strconv"
	"strings"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/", handleWebhook)

	port := 8080
	fmt.Printf("Server listening on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	signature := gjson.Get(string(body), "0.transaction.signatures.0").String()
	token := gjson.Get(string(body), "0.meta.postTokenBalances.1.mint").String()
	amount := gjson.Get(string(body), "0.meta.preTokenBalances.0.uiTokenAmount.uiAmount").Float()

	for i := 0; i < 14; i++ {
		instruction := gjson.Get(string(body), "0.meta.logMessages." + strconv.Itoa(i)).String()
		//fmt.Println(instruction)

		if instruction == "Program log: Instruction: LendingAccountRepay" {
			fmt.Println(signature)
			fmt.Println(string(body))

			amounttrue := gjson.Get(string(body), "0.meta.logMessages." + strconv.Itoa(i+4)).String()

			amounttrueint := 0

			if strings.Contains(amounttrue, "Program log: Balance increase: ") {
				amounttrue = strings.Replace(amounttrue, "Program log: Balance increase: ", "", -1)
				amounttrue = strings.Replace(amounttrue, " (type: RepayOnly)", "", -1)
				fmt.Println("geez", amounttrue)
				amounttrueint, _ = strconv.Atoi(amounttrue)
			}
			if strings.Contains(amounttrue, "Program log: deposit_spl_transfer: amount: ") {
				amounttrue = strings.Replace(amounttrue, "Program log: deposit_spl_transfer: amount: ", "", -1)
				index := strings.Index(amounttrue, "from")
				amounttrueint, _ = strconv.Atoi(amounttrue[0:index-2])
			}

			fmt.Println(amount, float64(amounttrueint) / 1e6)

			if token == "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" && float64(amounttrueint) / 1e6 >= 1000.0 {

				var username = "Marginfi monitor"
				var content = fmt.Sprintf("%.2f", float64(amounttrueint)/1e6) + " $USDC repaid\n" + "Signature: " + signature
				var url = ""

				message := discordwebhook.Message{
					Username: &username,
					Content:  &content,
				}

				err = discordwebhook.SendMessage(url, message)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println("Received data:", string(body))
				if token == "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" {
					fmt.Println("USDC")
				} else {
					fmt.Println(token)
				}
				fmt.Println(amount, float64(amounttrueint)/1e6)
				fmt.Println(signature)
				fmt.Println()

				break
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Received webhook data")
}

