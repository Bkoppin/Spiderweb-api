package neoModels

import neo "api/internal/app/neo4j"

type User struct {
	neo.NeoBaseModel[User]
	Username string   `node:"username" json:"username,omitempty"`
	UserID   int64    `node:"userID" json:"userID,omitempty"`
	ID       string   `node:"id" json:"id,omitempty"`
	Worlds   []*World `rel:"OWNS,->" json:"worlds,omitempty"`
}

type World struct {
	neo.NeoBaseModel[World]
	ID          string       `node:"id" json:"id,omitempty"`
	Name        string       `node:"name" json:"name,omitempty"`
	Type        string       `node:"type" json:"type,omitempty"`
	Description string       `node:"description" json:"description,omitempty"`
	Continents  []*Continent `rel:"HAS,->" json:"continents,omitempty"`
	Oceans      []*Ocean     `rel:"HAS,->" json:"oceans,omitempty"`
}

type Continent struct {
	neo.NeoBaseModel[Continent]
	ID          string  `node:"id" json:"id,omitempty"`
	Name        string  `node:"name" json:"name,omitempty"`
	Description string  `node:"description" json:"description,omitempty"`
	Type        string  `node:"type" json:"type,omitempty"`
	Zones       []*Zone `rel:"HAS,->" json:"zones,omitempty"`
}

type Ocean struct {
	neo.NeoBaseModel[Ocean]
	ID          string `node:"id" json:"id,omitempty"`
	Name        string `node:"name" json:"name,omitempty"`
	Description string `node:"description" json:"description,omitempty"`
}

type Zone struct {
	neo.NeoBaseModel[Zone]
	ID          string      `node:"id" json:"id,omitempty"`
	Name        string      `node:"name" json:"name,omitempty"`
	Type        string      `node:"type" json:"type,omitempty"`
	Description string      `node:"description" json:"description,omitempty"`
	Locations   []*Location `rel:"HAS,->" json:"locations,omitempty"`
	Cities      []*City     `rel:"HAS,->" json:"cities,omitempty"`
	Biome       string      `node:"biome" json:"biome,omitempty"`
}

type Location struct {
	neo.NeoBaseModel[Location]
	ID          string `node:"id" json:"id,omitempty"`
	Name        string `node:"name" json:"name,omitempty"`
	Type        string `node:"type" json:"type,omitempty"`
	Description string `node:"description" json:"description,omitempty"`
}

type City struct {
	neo.NeoBaseModel[City]
	ID          string `node:"id" json:"id,omitempty"`
	Name        string `node:"name" json:"name,omitempty"`
	Type        string `node:"type" json:"type,omitempty"`
	Description string `node:"description" json:"description,omitempty"`
	Capital     bool   `node:"capital" json:"capital,omitempty"`
}
