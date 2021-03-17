package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// Settings
var startingCash = 2000
const passGo = 200
var landGo = passGo * 2

// Global Variables
var properties []Property
var players []Player


// Property

type Property struct {
	name      string
	color     string
	code      string
	cost      int
	owner     string
	houses    int
	houseCost int
	rent      []int
	mortgaged bool
}

func loadProperties() []Property {
	csvFile, ok := os.Open("Properties.csv")
	if ok != nil {
		fmt.Println("Could not open Properties file")
	}
	
	csvLines, ok := csv.NewReader(csvFile).ReadAll()
	if ok != nil {
		fmt.Println("Could not read csv")
	}

	var newProperties []Property
	for _, line := range csvLines[1:] {
		cost, _ := strconv.Atoi(line[3])
		houseCost, _ := strconv.Atoi(line[5])
		rent := strings.Split(line[4], " ")
		var rentAmounts []int
		for _, rentString := range rent {
			rentCost, _ := strconv.Atoi(rentString)
			rentAmounts = append(rentAmounts, rentCost)
		}

		newProperty := Property {
			name: line[0],
			color: line[1],
			code: line[2],
			cost: cost,
			houses: 0,
			houseCost: houseCost,
			rent: rentAmounts,
			mortgaged: false,
		}

		newProperties = append(newProperties, newProperty)
	}

	return newProperties
}

func findProperty(code string) (*Property, error) {
	code = strings.ToUpper(code)
	for index, property := range properties {
		if code == property.code {
			return &properties[index], nil
		}
	}

	return &Property{}, errors.New("Could not find property code \"" + code + "\"")
}

func displayProperty(code string) {
	property, err := findProperty(code)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Name: %v\nColor: %v\nCode: %v\nCost: %v\nOwner: %v\nHouses: %v\nHouse Cost: %v\nRent: %v\nMortgaged: %v\n", property.name, property.color, property.code, property.cost, property.owner, property.houses, property.houseCost, property.rent, property.mortgaged)
}

func colorProperties(color string) []Property {
	var monoProperties []Property
	for _, property := range properties {
		if property.color == color {
			monoProperties = append(monoProperties, property)
		}
	}

	return monoProperties
}

func refreshMonopolies() {
	for index := range players {
		players[index].monopolies = nil
	}

	colors := []string{properties[0].color}
	for _, property := range properties[1:] {
		if property.color != colors[len(colors) - 1] {
			colors = append(colors, property.color)
		}
	}

	for _, color := range colors {
		monoProperties := colorProperties(color)
		owner, err := findPlayer(monoProperties[0].owner)
		if err != nil || monoProperties[0].mortgaged {
			continue
		}
		monopoly := true
		OwnerCheck:
		for _, property := range monoProperties[1:] {
			if property.owner != owner.name || property.mortgaged{
				monopoly = false
				break OwnerCheck
			}
		}

		if monopoly {
			owner.monopolies = append(owner.monopolies, color)
		}
	}
}


// Player

type Player struct {
	name 	   string
	cash 	   int
	properties []Property
	monopolies []string
}

func createPlayers() []Player{
	var playerCount int
	fmt.Print("Enter player count\n> ")
	_, err := fmt.Scan(&playerCount)

	if err != nil {
		return []Player{}
	}

	var createdPlayers []Player
	for i := 0; i < playerCount; {
		var name string
		fmt.Printf("Enter player #%v name:  ", i+1)
		_, err = fmt.Scan(&name)
		if err != nil {
			fmt.Println("Enter a valid name")
			continue
		}
		
		name = strings.ToLower(name)
		newPlayer := Player{
			name: name,
			cash: startingCash,
		}
		createdPlayers = append(createdPlayers, newPlayer)

		i++
	}

	return createdPlayers
}

func findPlayer(name string) (*Player, error) {
	name = strings.ToLower(name)
	for index, player := range players {
		if player.name == name {
			return &players[index], nil
		}
	}
	return &Player{}, errors.New("No player \"" + name + "\" found")
}

func playerColorCount(color string, player Player) int {
	var count int
	for _, property := range player.properties {
		if property.color == color {
			count++
		}
	}

	return count
}

func hasMonopoly(color string, player Player) (bool, bool) {
	var mortgaged bool
	for _, monopoly := range player.monopolies {
		if color == monopoly {
			for _, property := range colorProperties(color) {
				if property.mortgaged {
					mortgaged = true
					break
				}
			}
			return true, mortgaged
		}
	}
	return false, false
}

func displayPlayersCash() {
	var playerNames string
	var playerCash string
	for _, player := range players {
		playerNames += strings.Title(player.name) + strings.Repeat(" ", 15 - len(player.name))
		cashS := strconv.Itoa(player.cash)
		playerCash += cashS + strings.Repeat(" ", 15 - len(cashS))
	}
	fmt.Println(strings.Repeat("-", len(players) * 15))
	fmt.Println(playerNames)
	fmt.Print(playerCash, "\n\n")
}


// Other

func notEnough(playerCash, transferAmount int) {
	fmt.Printf("Not enough cash (Need $%v more)\n", transferAmount - playerCash)
}

func displayPlayer(player Player) {
	fmt.Println("Name: ", player.name)
	fmt.Println("Cash: ", player.cash)
	fmt.Println("Properties: ")
	for _, property := range player.properties {
		fmt.Println("\t", property.name, property.code)
	}
	fmt.Println("Monopolies: ")
	for _, monopoly := range player.monopolies {
		fmt.Println("\t", monopoly)
	}

}

/******


 Commands


 *******/
func processCommands() {
	var input string
	fmt.Print("\n\n> ")
	_, err := fmt.Scan(&input)
	if err != nil {
		return
	}

	input = strings.ToLower(input)

	commands := map[string]func() {
		"add": addCash,
		"bill": bill,
		"buy": buyProperty,
		"go": addGoPass,
		"lgo": addGoLand,
		"house": house,
		"mort": mortgage,
		"unmort": unmortgage,
		"pay": pay,
		"plr": playerValues,
		"prop": showProperty,
		"rem": removeCash,
		"rent": rent,
		"ride": ride,
		"roll": roll,
		"sell": sellProperty,
	}

	if input == "help" {
		for k := range commands {
			fmt.Println(k)
		}
		return
	}

	command, ok := commands[input]
	if ok {
		command()
	} else {
		fmt.Println("Enter valid command")
	}

}

func changeCash(name string, amount int) {

	player, err := findPlayer(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	player.cash += amount
	if amount < 0 {
		fmt.Printf("Removed %v from %v's account\n", -amount, name)
		return
	}
	fmt.Printf("Added %v to %v's account\n", amount, name)
}

func addCash() {
	fmt.Print("<name> <amount>\n> ")
	var name string
	var amount int
	_, ok := fmt.Scan(&name)
	if ok != nil {
		fmt.Println("Invalid input")
		return
	}

	_, ok = fmt.Scan(&amount)
	if ok != nil || amount < 0 {
		fmt.Println("Invalid amount")
		return
	}

	changeCash(name, amount)
}

func bill() {
	var payerName string
	var utilityCode string
	var roll int
	fmt.Printf("<payer> <utility code> <roll>\n> ")
	_, ok := fmt.Scan(&payerName)
	if ok != nil {
		fmt.Println("Invalid payer name")
		return
	}

	_, ok = fmt.Scan(&utilityCode)
	if ok != nil {
		fmt.Println("Invalid utility code")
		return
	}

	_, ok = fmt.Scan(&roll)
	if ok != nil {
		fmt.Println("Invalid roll")
		return
	}


	payer, err := findPlayer(payerName)
	if err != nil {
		fmt.Println(err)
		return
	}

	utility, err := findProperty(utilityCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	owner, err := findPlayer(utility.owner)
	if err != nil {
		fmt.Println(err)
		return
	}


	billAmount := roll * 4
	if mono, _ := hasMonopoly(utility.color, *owner); mono {
		billAmount = roll * 10
	}

	if payer.cash < billAmount {
		notEnough(payer.cash, billAmount)
		return
	}

	payer.cash -= billAmount
	owner.cash += billAmount
	fmt.Printf("%v was billed %v for %v's %v\n", payer.name, billAmount, owner.name, utility.name)
}

func buyProperty() {
	fmt.Printf("<buyer> <property code>\n> ")
	var buyer string
	var propertyCode string
	_, ok := fmt.Scan(&buyer)
	if ok != nil {
		fmt.Println("Invalid buyer")
		return
	}

	_, ok = fmt.Scan(&propertyCode)
	if ok != nil {
		fmt.Println("Invalid property code")
		return
	}


	player, err := findPlayer(buyer)
	if err != nil {
		fmt.Println(err)
		return
	}

	property, err := findProperty(propertyCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	
	if len(property.owner) > 0 {
		fmt.Printf("%v already owned by %v\n", property.code, property.owner)
		return
	}
	if player.cash < property.cost {
		notEnough(player.cash, property.cost)
		return
	}
	
	property.owner = player.name
	player.cash -= property.cost
	player.properties = append(player.properties, *property)

	fmt.Printf("%v bought %v for %v\n", player.name, property.name, property.cost)

	refreshMonopolies()
}

func addGo(land bool) {
	var name string
	fmt.Printf("<player>\n> ")
	_, ok := fmt.Scan(&name)
	if ok != nil {
		fmt.Println("Invalid player")
		return
	}

	player, err := findPlayer(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	if land {
		player.cash += landGo
		fmt.Printf("%v received %v for landing on Go\n", player.name, landGo)
		return
	}

	player.cash += passGo
	fmt.Printf("%v received %v for passing Go\n", player.name, passGo)
}

func addGoPass() {
	addGo(false)
}

func addGoLand() {
	addGo(true)
}

func house() {
	var propertyCode string
	var action string
	var houseCount int
	fmt.Printf("<property code> <buy|sell> <house count>\n> ")
	_, ok := fmt.Scan(&propertyCode)
	if ok != nil {
		fmt.Println("Invalid code")
		return
	}

	_, ok = fmt.Scan(&action)
	if ok != nil {
		fmt.Println("Invalid action")
		return
	}

	_, ok = fmt.Scan(&houseCount)
	if ok != nil {
		fmt.Println("Invalid house amount")
		return
	}

	property, err := findProperty(propertyCode)
	if err != nil {
		fmt.Println(err)
		return
	}
	if strings.HasPrefix(property.code, "RL") || strings.HasPrefix(property.code, "U") {
		fmt.Printf("Cannot build houses on %v\n", property.color)
		return
	}

	player, err := findPlayer(property.owner)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Check if has monopoly and not mortgaged
	if mono, mort := hasMonopoly(property.color, *player); mono && mort{
		fmt.Printf("%v does not have the %v monopoly\n", player.name, property.color)
		return
	}

	switch action {
	case "buy":
		if property.houses + houseCount > 5 {
			fmt.Println("Too many houses")
			return
		}
		totalCost := houseCount * property.houseCost
		if player.cash < totalCost {
			notEnough(player.cash, totalCost)
			return
		}
		player.cash -= totalCost
		property.houses += houseCount
		fmt.Printf("%v bought %v houses on %v for %v\n", player.name, houseCount, property.name, totalCost)
	case "sell":
		if houseCount > property.houses {
			fmt.Printf("Property only has %v houses\n", property.houses)
		}
		player.cash += (houseCount * property.houseCost) / 2
		property.houses -= houseCount
	default:
		fmt.Println("Invalid action")
	}
}

func mortgage() {
	var propertyCode string
	fmt.Printf("<property code>\n> ")
	_, ok := fmt.Scan(&propertyCode)
	if ok != nil {
		fmt.Println("Invalid property code")
		return
	}


	property, err := findProperty(propertyCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	owner, err := findPlayer(property.owner)
	if err != nil {
		fmt.Println(err)
		return
	}

	if property.houses > 0 {
		fmt.Println("Cannot mortgage property with houses")
		return
	}
	if property.mortgaged {
		fmt.Println("Property already mortgaged")
		return
	}

	mortgageTotal := property.cost / 2
	owner.cash += mortgageTotal
	property.mortgaged = true
	fmt.Printf("%v mortgaged %v for %v\n", owner.name, property.name, mortgageTotal)

	refreshMonopolies()
}

func unmortgage() {
	var propertyCode string
	fmt.Printf("<property code>\n> ")
	_, ok := fmt.Scan(&propertyCode)
	if ok != nil {
		fmt.Println("Invalid property code")
		return
	}


	property, err := findProperty(propertyCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	owner, err := findPlayer(property.owner)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !property.mortgaged {
		fmt.Println("Property already mortgaged")
		return
	}

	unmortgageCost := property.cost / 2 + property.cost / 20
	if unmortgageCost > owner.cash {
		notEnough(owner.cash, unmortgageCost)
	}
	owner.cash -= unmortgageCost
	property.mortgaged = false
	fmt.Printf("%v mortgaged %v for %v\n", owner.name, property.name, unmortgageCost)
	refreshMonopolies()
}

func pay() {
	var senderName string
	var amount int
	var receiverName string
	fmt.Printf("<sender> <amount> <receiver>\n> ")
	_, ok := fmt.Scan(&senderName)
	if ok != nil {
		fmt.Println("Invalid sender name")
		return
	}

	_, ok = fmt.Scan(&amount)
	if ok != nil {
		fmt.Println("Invalid amount")
		return
	}

	_, ok = fmt.Scan(&receiverName)
	if ok != nil {
		fmt.Println("Invalid receiver name")
		return
	}


	sender, err := findPlayer(senderName)
	if err != nil {
		fmt.Println(err)
		return
	}

	receiver, err := findPlayer(receiverName)
	if err != nil {
		fmt.Println(err)
		return
	}


	if sender.cash < amount {
		fmt.Printf("%v only has %v cash\n", sender.name, sender.cash)
		return
	}

	sender.cash -= amount
	receiver.cash += amount
	fmt.Printf("%v payed %v %v\n", sender.name, receiver.name, amount)
}

func playerValues() {
	var name string
	var option string
	fmt.Printf("<player name> <all|cash|prop|monos>\n> ")
	_, ok := fmt.Scan(&name)
	if ok != nil {
		fmt.Println("Invalid name")
		return
	}
	
	_, ok = fmt.Scan(&option)
	if ok != nil {
		fmt.Println("Invalid option")
		return
	}
	
	player, err := findPlayer(name)
	if err != nil {
		fmt.Println(err)
		return
	}
	
	switch option {
	case "all":
		displayPlayer(*player)
	case "cash":
		fmt.Println(player.cash)
	case "prop":
		for _, property := range player.properties {
			fmt.Println(property.name, property.code)
		}
	case "monos":
		for _, monopoly := range player.monopolies {
			fmt.Println(monopoly)
		}
	}
}

func showProperty() {
	var code string
	fmt.Printf("<property code>\n> ")
	_, ok := fmt.Scan(&code)
	if ok != nil {
		fmt.Println("Invalid code")
		return
	}

	displayProperty(code)
}

func removeCash() {
	fmt.Print("<name> <amount>\n> ")
	var name string
	var amount int
	_, ok := fmt.Scan(&name)
	if ok != nil {
		fmt.Println("Invalid input")
		return
	}

	_, ok = fmt.Scan(&amount)
	if ok != nil || amount < 0 {
		fmt.Println("Invalid amount")
		return
	}

	changeCash(name, -amount)
}

func rent() {
	var renterName string
	var propertyCode string
	fmt.Printf("<renter> <property code>\n> ")
	_, ok := fmt.Scan(&renterName)
	if ok != nil {
		fmt.Println("Invalid buyer")
		return
	}

	_, ok = fmt.Scan(&propertyCode)
	if ok != nil {
		fmt.Println("Invalid property code")
		return
	}


	renter, err := findPlayer(renterName)
	if err != nil {
		fmt.Println(err)
		return
	}

	property, err := findProperty(propertyCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	owner, err := findPlayer(property.owner)
	if err != nil {
		fmt.Println(err)
		return
	}


	rentAmount := property.rent[property.houses]
	if mono, _ := hasMonopoly(property.color, *owner); mono && property.houses == 0 {
		rentAmount = property.rent[0] * 2
	}

	if renter.cash < rentAmount {
		notEnough(renter.cash, rentAmount)
	}
	renter.cash -= rentAmount
	owner.cash += rentAmount

	fmt.Printf("%v rented %v's %v for %v\n", renter.name, owner.name, property.name, rentAmount)

}

func ride() {
	var riderName string
	var railroadCode string
	fmt.Printf("<rider> <railroad code>\n> ")
	_, ok := fmt.Scan(&riderName)
	if ok != nil {
		fmt.Println("Invalid rider")
		return
	}

	_, ok = fmt.Scan(&railroadCode)
	if ok != nil {
		fmt.Println("Invalid railroad code")
		return
	}


	rider, err := findPlayer(riderName)
	if err != nil {
		fmt.Println(err)
		return
	}

	railroad, err := findProperty(railroadCode)
	if err != nil {
		fmt.Println(err)
		return
	}

	owner, err := findPlayer(railroad.owner)
	if err != nil {
		fmt.Println(err)
		return
	}


	var rideCost = 25
	for i := 1; i < playerColorCount(railroad.color, *owner); i++ {
		rideCost *= 2
	}
	if rider.cash < rideCost {
		notEnough(rider.cash, rideCost)
		return
	}

	rider.cash -= rideCost
	owner.cash += rideCost
	fmt.Printf("%v rode %v's %v for %v\n", rider.name, owner.name, railroad.name, rideCost)
}

func roll() {
	rand.Seed(time.Now().UnixNano())
	die1 := rand.Intn(6) + 1
	die2 := rand.Intn(6) + 1
	fmt.Printf("%v %v (%v)\n", die1, die2, die1 + die2)
}

func sellProperty() {
	var propertyCount int
	fmt.Printf("<property sell count>\n> ")
	_, ok := fmt.Scan(&propertyCount)
	if ok != nil {
		fmt.Println("Invalid property count")
		return
	}
	
	var sellerName string
	var buyerName string
	var sellPrice int
	var propertyCodeList []string
	fmt.Printf("<seller> <buyer> <sell price> <property codes>\n> ")
	_, ok = fmt.Scan(&sellerName)
	if ok != nil {
		fmt.Println("Invalid seller name")
		return
	}

	_, ok = fmt.Scan(&buyerName)
	if ok != nil {
		fmt.Println("Invalid buyer name")
		return
	}

	_, ok = fmt.Scan(&sellPrice)
	if ok != nil {
		fmt.Println("Invalid sell price")
		return
	}

	for i := 0; i < propertyCount; i++ {
		var propertyCode string
		_, ok = fmt.Scan(&propertyCode)
		if ok != nil {
			fmt.Println("Invalid property code")
			return
		}
		
		propertyCodeList = append(propertyCodeList, propertyCode)
	}


	seller, err := findPlayer(sellerName)
	if err != nil {
		fmt.Println(err)
		return
	}

	buyer, err := findPlayer(buyerName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if buyer.cash < sellPrice {
		fmt.Printf("%v only has %v cash\n", buyer.name, buyer.cash)
	}

	for _, code := range propertyCodeList {
		property, err := findProperty(code)
		if err != nil {
			fmt.Println(err)
			continue
		}

		property.owner = buyer.name
		buyer.properties = append(buyer.properties, *property)
	}

	oldPropertyList := seller.properties
	seller.properties = nil
	for _, propertyValue := range oldPropertyList {
		soldProperty := false
		for _, code := range propertyCodeList {
			if strings.ToUpper(code) == propertyValue.code {
				soldProperty = true
				break
			}
		}

		if !soldProperty {
			property, err := findProperty(propertyValue.code)
			if err != nil {
				fmt.Println("Invalid property")
				continue
			}
			seller.properties = append(seller.properties, *property)
		}
	}

	buyer.cash -= sellPrice
	seller.cash += sellPrice

	fmt.Printf("%v sold %v to %v for %v\n", seller.name, propertyCodeList, buyer.name, sellPrice)
	refreshMonopolies()
}


func main() {
	properties = loadProperties()
	players = createPlayers()
	for {
		displayPlayersCash()
		processCommands()
	}
}
