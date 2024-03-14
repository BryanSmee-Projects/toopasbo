package transformers

import (
	"crypto/rand"
	"math/big"
)

var animalsPerTemperature = map[string][]string{
	"freezing": {"Polar Bear",
		"Emperor Penguin",
		"Arctic Fox",
		"Snow Leopard",
		"Narwhal",
		"Walrus",
		"Musk Ox",
		"Arctic Hare",
		"Beluga Whale",
		"Canadian Lynx",
		"Icefish",
		"Weddell Seal",
		"Arctic Tern",
		"Snowy Owl",
		"Antarctic Krill",
		"Puffin",
		"Svalbard Reindeer",
		"Snow Petrel",
		"Greenland Shark",
		"Antarctic Sponge"},
	"cold": {
		"Moose",
		"Red Fox",
		"European Badger",
		"Brown Bear",
		"Elk",
		"Gray Wolf",
		"Bearded Seal",
		"Siberian Tiger",
		"Reindeer",
		"Snowy Owl",
		"Norwegian Lemming",
		"Atlantic Puffin",
		"Wolverine",
		"Dall Sheep",
		"Steller's Sea Eagle",
		"Harp Seal",
		"Mountain Goat",
		"Siberian Musk Deer",
		"Sea Otter",
		"Atlantic Salmon",
	},
	"mild": {
		"Common Toad",
		"Scottish Wildcat",
		"Eurasian Otter",
		"Marsh Tit",
		"European Green Woodpecker",
		"Fire Salamander",
		"European Badger",
		"Barn Owl",
		"Red Squirrel",
		"Kingfisher",
		"European Hedgehog",
		"Red Deer",
		"Wild Boar",
		"Eurasian Lynx",
		"Black Grouse",
		"Pine Marten",
		"Golden Eagle",
		"Iberian Lynx",
		"Common Frog",
		"Roe Deer",
	},
	"warm": {
		"American Black Bear",
		"Coyote",
		"White-tailed Deer",
		"Bobcat",
		"North American Beaver",
		"Eastern Gray Squirrel",
		"Bald Eagle",
		"Red-tailed Hawk",
		"Virginia Opossum",
		"Raccoon",
		"American Alligator",
		"Ruby-throated Hummingbird",
		"Eastern Cottontail",
		"Wood Duck",
		"American Bullfrog",
		"Eastern Bluebird",
		"Northern Cardinal",
		"American Goldfinch",
		"Copperhead Snake",
		"Box Turtle",
	},
	"nice": {
		"African Leopard",
		"Gibbon",
		"Malayan Tapir",
		"Clouded Leopard",
		"Asian Elephant",
		"Sun Bear",
		"Proboscis Monkey",
		"Green Anaconda",
		"Aye-aye",
		"Philippine Eagle",
		"Jaguar",
		"African Elephant",
		"Chimpanzee",
		"Bengal Tiger",
		"Orangutan",
		"Red Panda",
		"Giant Anteater",
		"Sloth Bear",
		"Komodo Dragon",
		"Capybara",
	},
	"hot": {
		"Lion",
		"Giraffe",
		"Hippopotamus",
		"Rhinoceros",
		"African Buffalo",
		"Cheetah",
		"Gorilla",
		"Zebra",
		"Ostrich",
		"Nile Crocodile",
		"Fennec Fox",
		"Meerkat",
		"Galápagos Tortoise",
		"Komodo Dragon",
		"Mandrill",
		"Okapi",
		"Bonobo",
		"Lemur",
		"Flamingo",
		"Caribbean Reef Shark",
	},
}

func temperatureToCategory(temperature int) string {
	switch {
	case temperature <= 0:
		return "freezing"
	case temperature < 8:
		return "cold"
	case temperature < 12:
		return "mild"
	case temperature < 20:
		return "warm"
	case temperature < 26:
		return "nice"
	default:
		return "hot"
	}
}

func GetAnimalsByTemperature(temperature int) string {
	animalSlice := animalsPerTemperature[temperatureToCategory(temperature)]
	randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(animalSlice))))
	return animalSlice[randomIndex.Int64()]
}
