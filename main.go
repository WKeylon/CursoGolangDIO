package main

import "fmt"

func main() {
	temperaturaKelvin := 0.0
	fmt.Print("Digite a Temperatura em Kelvin: ")
	fmt.Scanln(&temperaturaKelvin)
	temperaturaCelcius := (temperaturaKelvin - 273.15)
	fmt.Printf("A Temperatura em Celcius é de: %.2f ", temperaturaCelcius)
}
