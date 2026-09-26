package world

import (
	"fmt"
	"log"
)

type MobHandle struct {
	Id uint32
	Generation uint32
}

type MobIndex struct {
	index uint32
	generation uint32
}

type MobArray struct {
	freeHandles []MobHandle
	idToIndex []MobIndex // Sparse set
	indexToId []uint32 // Dense set, parallel to data
	data []Mob // Dense set, data
}

func MobArrayInit() MobArray {
	return MobArray {
		freeHandles: make([]MobHandle, 0, 1),
		idToIndex: make([]MobIndex, 0, 1),
		indexToId: make([]uint32, 0, 1),
		data: make([]Mob, 0, 1),
	}
}

func (array *MobArray) GetIfExists(handle MobHandle) (*Mob, bool) {
	if int(handle.Id) > len(array.idToIndex) - 1 {
		log.Printf("Warm - Tried to get Mob with ID %d which is greater than the list of IDs which has length %d", handle.Id, len(array.idToIndex))
		return nil, false
	}
	if array.idToIndex[handle.Id].generation != handle.Generation {
		return nil, false
	}

	index := array.idToIndex[handle.Id].index
	return &array.data[index], true
}

func (array *MobArray) Get(handle MobHandle) *Mob {
	mob, exists := array.GetIfExists(handle)
	if !exists {
		panic(fmt.Sprintf("Tried to get Mob with handle %d:%d, but it does not exist.", handle.Id, handle.Generation))
	}

	return mob
}

func (array *MobArray) Push(mob Mob) MobHandle {
	// Determine handle
	var handle MobHandle
	if len(array.freeHandles) != 0 {
		handle = array.freeHandles[0]
		array.freeHandles = array.freeHandles[1:]
	} else {
		// Since there's no free handles, we have to make a new handle and index
		handle = MobHandle {
			Id: uint32(len(array.idToIndex)),
			Generation: 0,
		}
		// Note: we're just allocating this element here, and it will be
		// populated by the statement below
		array.idToIndex = append(array.idToIndex, MobIndex {})
	}

	// Set ID to index
	array.idToIndex[handle.Id] = MobIndex {
		index: uint32(len(array.data)),
		generation: handle.Generation,
	}

	// Add data
	mob.Handle = handle
	array.data = append(array.data, mob)
	array.indexToId = append(array.indexToId, handle.Id)

	return handle
}

func (array *MobArray) Remove(handle MobHandle) {
	// Check that it's a valid handle
	if int(handle.Id) > len(array.idToIndex) {
		panic(fmt.Sprintf("Tried to remove Mob with ID %d which is greater than the lsit of IDs which has length %d", handle.Id, len(array.idToIndex)))
	}
	if array.idToIndex[handle.Id].generation != handle.Generation {
		panic(fmt.Sprintf("Tried to move Mob with handle %d:%d but the generation does not match the current generation %d",
			handle.Id,
			handle.Generation,
			array.idToIndex[handle.Id].generation),
		)
	}

	// Invalidate the current index->ID entry by incrementing the generation
	array.idToIndex[handle.Id].generation++

	// Add the handle to the list of free handles
	array.freeHandles = append(array.freeHandles, MobHandle {
		Id: handle.Id,
		Generation: handle.Generation + 1,
	})

	// Determine the index and ID of the last element
	lastIndex := uint32(len(array.data) - 1)
	lastId := array.indexToId[lastIndex]

	// Redirect the last element's ID to point to the removed index
	index := array.idToIndex[handle.Id].index
	array.idToIndex[lastId].index = index

	// Move the last index into the removed slot
	array.data[index] = array.data[lastIndex]
	array.indexToId[index] = array.indexToId[lastIndex]

	// Remove the last element from the array
	array.data = array.data[:lastIndex]
	array.indexToId = array.indexToId[:lastIndex]
}
