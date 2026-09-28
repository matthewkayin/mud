package api

import (
	"log"
	"mud/world"
	"net/http"
	"encoding/json"
)

func (apiState *ApiState) HandleWorldGetItem(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Invoked GET /api/world/items")

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	err := json.NewEncoder(writer).Encode(world.ITEM_DATA)
	if err != nil {
		http.Error(writer, "Error serializing item data", http.StatusInternalServerError)
		return
	}
}

func (apiState *ApiState) HandleWorldGetSpell(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Invoked GET /api/world/spells")

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	err := json.NewEncoder(writer).Encode(world.SPELL_DATA)
	if err != nil {
		http.Error(writer, "Error serializing spell data", http.StatusInternalServerError)
		return
	}
}
