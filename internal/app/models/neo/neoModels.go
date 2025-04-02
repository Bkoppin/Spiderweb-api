package neoModels

type User struct {
    Username string  `node:"username" json:"username"`
    UserID   int64   `node:"userID" json:"userID"`
    Worlds   []*World `rel:"World,->" json:"worlds"`
}

type World struct {
    ID          string           `node:"id" json:"id"`
    Name        string           `node:"name" json:"name"`
    Type        string           `node:"type" json:"type"`
    Description string           `node:"description" json:"description"`
    Continents  []*Continent     `rel:"Continent,->" json:"continents"` // Use []*Continent
    Oceans      []*Ocean         `rel:"Ocean,->" json:"oceans"`         // Use []*Ocean
}

type Continent struct {
    ID          string       `node:"id" json:"id"`
    Name        string       `node:"name" json:"name"`
    Description string       `node:"description" json:"description"`
    Type        string       `node:"type" json:"type"`
    Zones       []*Zone      `rel:"Zone,->" json:"zones"` // Use []*Zone
}

type Ocean struct {
    ID          string `node:"id" json:"id"`
    Name        string `node:"name" json:"name"`
    Description string `node:"description" json:"description"`
}

type Zone struct {
    ID          string       `node:"id" json:"id"`
    Name        string       `node:"name" json:"name"`
    Type        string       `node:"type" json:"type"`
    Description string       `node:"description" json:"description"`
    Locations   []*Location  `rel:"Location,->" json:"locations"` // Use []*Location
    Cities      []*City      `rel:"City,->" json:"cities"`        // Use []*City
    Biome       string       `node:"biome" json:"biome"`
}

type Location struct {
    ID          string `node:"id" json:"id"`
    Name        string `node:"name" json:"name"`
    Type        string `node:"type" json:"type"`
    Description string `node:"description" json:"description"`
}

type City struct {
    ID          string `node:"id" json:"id"`
    Name        string `node:"name" json:"name"`
    Type        string `node:"type" json:"type"`
    Description string `node:"description" json:"description"`
    Capital     bool   `node:"capital" json:"capital"`
}


