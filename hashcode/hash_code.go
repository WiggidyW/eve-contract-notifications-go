package hashcode

type HashCode string
type HashCodeSet map[HashCode]struct{}

func HashCodeSetNew() HashCodeSet {
	return make(HashCodeSet, 0)
}

func HashCodeSetWithCapacity(capacity int) HashCodeSet {
	return make(HashCodeSet, capacity)
}

func (h HashCodeSet) Add(hashCode HashCode) {
	h[hashCode] = struct{}{}
}

func (h HashCodeSet) Contains(hashCode HashCode) bool {
	_, ok := h[hashCode]
	return ok
}
