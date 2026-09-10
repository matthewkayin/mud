package game

import "fmt"

type MobHandle struct {
	id uint32
	generation uint32
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

func (handle *MobHandle) Equals(other MobHandle) bool {
	return handle.id == other.id && handle.generation == other.generation
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
	if int(handle.id) > len(array.idToIndex) {
		panic(fmt.Sprintf("Tried to get Mob with ID %d which is greater than the list of IDs which has length %d", handle.id, len(array.idToIndex)))
	}
	if array.idToIndex[handle.id].generation != handle.generation {
		return nil, false
	}

	index := array.idToIndex[handle.id].index
	return &array.data[index], true
}

func (array *MobArray) Get(handle MobHandle) *Mob {
	mob, exists := array.GetIfExists(handle)
	if !exists {
		panic(fmt.Sprintf("Tried to get Mob with handle %d:%d, but it does not exist.", handle.id, handle.generation))
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
			id: uint32(len(array.idToIndex)),
			generation: 0,
		}
		// Note: we're just allocating this element here, and it will be
		// populated by the statement below
		array.idToIndex = append(array.idToIndex, MobIndex {})
	}

	// Set ID to index
	array.idToIndex[handle.id] = MobIndex {
		index: uint32(len(array.data)),
		generation: handle.generation,
	}

	// Add data
	array.data = append(array.data, mob)
	array.indexToId = append(array.indexToId, handle.id)

	return handle
}

func (array *MobArray) GetHandleAtIndex(index uint32) MobHandle {
	if int(index) > len(array.data) {
		panic(fmt.Sprintf("Cannot remove index %d from MobArray with length %d", index, len(array.data)))
	}

	id := array.indexToId[index]
	return MobHandle {
		id: id,
		generation: array.idToIndex[id].generation,
	}
}

func (array *MobArray) RemoveAtIndex(index uint32) {
	mobHandle := array.GetHandleAtIndex(index)

	// Add this handle to the list of free handles
	array.freeHandles = append(array.freeHandles, MobHandle {
		id: mobHandle.id,
		generation: mobHandle.generation + 1,
	})

	// Determine the index and ID of the last element
	lastIndex := uint32(len(array.data) - 1)
	lastId := array.indexToId[lastIndex]

	// Redirect the last element's ID to point to the remove index
	array.idToIndex[lastId].index = index

	// Move last index into the removed index slot
	array.data[index] = array.data[lastIndex]
	array.idToIndex[index] = array.idToIndex[lastIndex]

	// Remove the last element from the array
	array.data = array.data[:lastIndex]
	array.idToIndex = array.idToIndex[:lastIndex]
}
