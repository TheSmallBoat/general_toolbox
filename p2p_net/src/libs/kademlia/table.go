package kademlia

import (
	"fmt"
	"sync"

	"awesomeProject/beacon/p2p_network/libs/cryptographic"
)

// Table represents a Kademlia routing table.
type Table struct {
	sync.RWMutex

	entries [cryptographic.SizePublicKey * 8][]cryptographic.ID
	self    cryptographic.ID
	size    int
}

// NewTable instantiates a new routing table whose XOR distance metric is defined with respect to some
// given ID.
func NewTable(self cryptographic.ID) *Table {
	table := &Table{self: self}

	if _, err := table.Update(self); err != nil {
		panic(err)
	}

	return table
}

// Self returns the ID which this routing table's XOR distance metric is defined with respect to.
func (t *Table) Self() cryptographic.ID {
	return t.self
}

// Bucket returns all IDs in the bucket where target resides within.
func (t *Table) Bucket(target cryptographic.PublicKey) []cryptographic.ID {
	t.RLock()
	defer t.RUnlock()

	return t.entries[t.getBucketIndex(target)]
}

// Update attempts to insert the target node/peer ID into this routing table. If the bucket it was expected
// to be inserted within is full, ErrBucketFull is returned. If the ID already exists in its respective routing
// table bucket, it is moved to the head of the bucket and false is returned. If the ID has yet to exist, it is
// appended to the head of its intended bucket and true is returned.
func (t *Table) Update(target cryptographic.ID) (bool, error) {
	if target.PubKey == cryptographic.ZeroPublicKey {
		return false, nil
	}

	t.Lock()
	defer t.Unlock()

	idx := t.getBucketIndex(target.PubKey)

	for i, id := range t.entries[idx] {
		if id.PubKey == target.PubKey { // Found the target ID already inside the routing table.
			t.entries[idx] = append(append([]cryptographic.ID{target}, t.entries[idx][:i]...), t.entries[idx][i+1:]...)
			return false, nil
		}
	}

	if len(t.entries[idx]) < BucketSize { // The bucket is not yet under full capacity.
		t.entries[idx] = append([]cryptographic.ID{target}, t.entries[idx]...)
		t.size++
		return true, nil
	}

	// The bucket is at full capacity. Return ErrBucketFull.

	return false, fmt.Errorf("cannot insert id %x into routing table: %v", target.PubKey, ErrBucketFull)
}

// Recorded returns true if target is already recorded in this routing table.
func (t *Table) Recorded(target cryptographic.PublicKey) bool {
	t.RLock()
	defer t.RUnlock()

	for _, id := range t.entries[t.getBucketIndex(target)] {
		if id.PubKey == target {
			return true
		}
	}

	return false
}

// Delete removes target from this routing table. It returns the id of the delted target and true if found, or
// a zero-value ID and false otherwise.
func (t *Table) Delete(target cryptographic.PublicKey) (cryptographic.ID, bool) {
	t.Lock()
	defer t.Unlock()

	idx := t.getBucketIndex(target)

	for i, id := range t.entries[idx] {
		if id.PubKey == target {
			t.entries[idx] = append(t.entries[idx][:i], t.entries[idx][i+1:]...)
			t.size--
			return id, true
		}
	}

	return cryptographic.ID{}, false
}

// DeleteByAddress removes the first occurrence of an id with target as its address from this routing table. It
// returns the id of the deleted target and true if found, or a zero-value ID and false otherwise.
func (t *Table) DeleteByAddress(target string) (cryptographic.ID, bool) {
	t.Lock()
	defer t.Unlock()

	for i, bucket := range t.entries {
		for j, id := range bucket {
			if id.Address == target {
				t.entries[i] = append(t.entries[i][:j], t.entries[i][j+1:]...)
				t.size--
				return id, true
			}
		}
	}

	return cryptographic.ID{}, false
}

// Peers returns BucketSize closest peer IDs to the ID which this routing table's distance metric is defined against.
func (t *Table) Peers() []cryptographic.ID {
	return t.FindClosest(t.self.PubKey, BucketSize)
}

// FindClosest returns the k closest peer IDs to target, and sorts them based on how close they are.
func (t *Table) FindClosest(target cryptographic.PublicKey, k int) []cryptographic.ID {
	var closest []cryptographic.ID

	f := func(bucket []cryptographic.ID) {
		for _, id := range bucket {
			if id.PubKey != target {
				closest = append(closest, id)
			}
		}
	}

	t.RLock()
	defer t.RUnlock()

	idx := t.getBucketIndex(target)

	f(t.entries[idx])

	for i := 1; len(closest) < k && (idx-i >= 0 || idx+i < len(t.entries)); i++ {
		if idx-i >= 0 {
			f(t.entries[idx-i])
		}

		if idx+i < len(t.entries) {
			f(t.entries[idx+i])
		}
	}

	closest = SortByDistance(target, closest)

	if len(closest) > k {
		closest = closest[:k]
	}

	return closest
}

// Entries returns all stored ids in this routing table.
func (t *Table) Entries() []cryptographic.ID {
	t.RLock()
	defer t.RUnlock()

	entries := make([]cryptographic.ID, 0, t.size)

	for _, bucket := range t.entries {
		for _, id := range bucket {
			entries = append(entries, id)
		}
	}

	return entries
}

// NumEntries returns the total amount of ids stored in this routing table.
func (t *Table) NumEntries() int {
	t.RLock()
	defer t.RUnlock()

	return t.size
}

func (t *Table) getBucketIndex(target cryptographic.PublicKey) int {
	l := PrefixLen(XOR(target[:], t.self.PubKey[:]))
	if l == cryptographic.SizePublicKey*8 {
		return l - 1
	}

	return l
}
