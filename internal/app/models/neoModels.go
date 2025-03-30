package models

type NeoUser struct {
	Username string  `json:"username"`
	UserID   int64  `json:"userID"`
	Worlds   []World `json:"worlds"`
}

type World struct {
	Name        string      `json:"name"`
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Continents  []Continent `json:"continents"`
	Oceans      []Ocean     `json:"oceans"`
}

type Continent struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Zones       []Zone `json:"zones"`
}

type Ocean struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Zone struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	Locations   []Location `json:"locations"`
	Cities      []City     `json:"cities"`
	Biome       string     `json:"biome"`
}

type Location struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type City struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Capital     bool   `json:"capital"`
}
