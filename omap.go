package bsony

type omapElement struct {
	Element

	// Following fields only used in building a Map
	k    string
	next *omapElement
	prev *omapElement
}

// An OM is an ordered map...
type OM struct {
	by    *Factory
	index map[string]*omapElement
	root  *omapElement // sentinel
}

// // XXX instantiate from *B
// func newOM(bx *B) *OM {
// 	om := &OM{bx: bx, root: Element{}}
// 	om.root.next = &om.root
// 	om.root.prev = &om.root
// 	om.index = make(map[string]*Element)
// 	return om
// }

// func (om *OM) insertAfter(e, after *Element) {
// 	next := after.next
// 	after.next = e
// 	e.prev = after
// 	e.next = next
// 	next.prev = e
// }

// func (om *OM) remove(e Element) {
// 	e.next.prev = e.prev
// 	e.prev.next = e.next
// 	delete(om.index, e.k)
// 	e.Release()
// 	// Avoid memory leaks
// }

// // Len ...
// func (om *OM) Len() int {
// 	return len(om.index)
// }

// func append(om *OM, e *Element) {
// 	// Any duplicate is first deleted
// 	om.Delete(e.k)
// 	switch {
// 	case om.head == nil:
// 		om.head = e
// 		om.tail = e
// 		om.index[e.k] = e
// 	default:
// 		e.prev = om.tail
// 		om.tail.next = e
// 		om.tail = e
// 		om.index[e.k] = e
// 	}
// }

// // Delete ... (is a NOP if it doesn't exist)
// func (om *OM) Delete(k string) {
// 	e, ok := om.index[k]
// 	if !ok {
// 		return
// 	}
// 	if om.head == e {
// 		om.head = e.next
// 	} else {

// 	}
// 	if om.tail == e {

// 	} else {

// 	}
// }

// // Add ...
// func (om *OM) Add(k string, v interface{}) *OM {
// 	// XXX
// 	// Dispatch based on type
// 	// When adding we should always update the length so that "d"
// 	// is always a valid BSON document and doesn't need to be "closed".
// 	return om
// }

// // AddDouble ...
// func (om *OM) AddDouble(k string, v float64) *OM {
// 	// 8 bytes for float64
// 	buf := om.bx.Pool.Get().([]byte)
// 	om.bx.Pool.Resize(buf, 8)
// 	writeFloat64(buf, 0, v)
// 	e := &Element{bx: om.bx, t: TypeDouble, data: buf, owns: true}
// 	return om
// }

// // AddString ...
// func (om *OM) AddString(k string, v string) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + length + string + null byte
// 	om.grow(7 + len(k) + len(v))
// 	offset = writeTypeAndKey(om.buf, offset, TypeString, k)
// 	writeString(om.buf, offset, v)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddDoc ...
// func (om *OM) AddDoc(k string, v *D) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + sub document length
// 	om.grow(len(k) + v.Len() + 2)
// 	offset = writeTypeAndKey(om.buf, offset, TypeEmbeddedDocument, k)
// 	copy(om.buf[offset:], v.buf)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddArray ...
// func (om *OM) AddArray(k string, v *A) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + array length
// 	om.grow(len(k) + v.Len() + 2)
// 	offset = writeTypeAndKey(om.buf, offset, TypeArray, k)
// 	copy(om.buf[offset:], v.om.buf)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddBinary ...
// func (om *OM) AddBinary(k string, v *Binary) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + length + subtype byte +
// 	// payload; payload for subtype 2 also adds payload length bytes
// 	dataSize := len(v.Payload)
// 	if v.Subtype == 2 {
// 		dataSize += 4
// 	}
// 	om.grow(7 + len(k) + dataSize)
// 	offset = writeTypeAndKey(om.buf, offset, TypeBinary, k)
// 	offset = writeInt32(om.buf, offset, int32(dataSize))
// 	om.buf[offset] = v.Subtype
// 	offset++
// 	if v.Subtype == 2 {
// 		offset = writeInt32(om.buf, offset, int32(len(v.Payload)))
// 	}
// 	copy(om.buf[offset:], v.Payload)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddUndefined ...
// func (om *OM) AddUndefined(k string) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte; no data bytes
// 	om.grow(2 + len(k))
// 	offset = writeTypeAndKey(om.buf, offset, TypeUndefined, k)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddOID ...
// func (om *OM) AddOID(k string, v OID) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + OID length (12)
// 	om.grow(len(k) + 14)
// 	offset = writeTypeAndKey(om.buf, offset, TypeObjectID, k)
// 	copy(om.buf[offset:], v[:])
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddBool ...
// func (om *OM) AddBool(k string, v bool) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + boolean byte
// 	om.grow(len(k) + 3)
// 	offset = writeTypeAndKey(om.buf, offset, TypeBoolean, k)
// 	if v {
// 		om.buf[offset] = 1
// 	} else {
// 		om.buf[offset] = 0
// 	}
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddDateTime ...
// func (om *OM) AddDateTime(k string, v time.Time) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	x := v.Unix()
// 	if x < minDateTimeSec || x > maxDateTimeSec {
// 		om.err = fmt.Errorf("time %v outside BSON DateTime range", v)
// 		return om
// 	}
// 	x *= 1000
// 	x += v.UnixNano() / 1000000
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + 8 for int64
// 	om.grow(10 + len(k))
// 	offset = writeTypeAndKey(om.buf, offset, TypeDateTime, k)
// 	writeInt64(om.buf, offset, x)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddNull ...
// func (om *OM) AddNull(k string) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte; no data bytes
// 	om.grow(2 + len(k))
// 	offset = writeTypeAndKey(om.buf, offset, TypeNull, k)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddRegex ...
// func (om *OM) AddRegex(k string, v Regex) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + len(pattern) + null byte +
// 	// len(flags) + null byte
// 	om.grow(3 + len(k) + len(v.Pattern) + len(v.Flags))
// 	offset = writeTypeAndKey(om.buf, offset, TypeInt32, k)
// 	offset = writeCString(om.buf, offset, v.Pattern)
// 	writeCString(om.buf, offset, v.Flags)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddDBPointer ...
// func (om *OM) AddDBPointer(k string, v DBPointer) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + length + string + null
// 	// byte + 12-byte ID
// 	om.grow(19 + len(k) + len(v.Ref))
// 	offset = writeTypeAndKey(om.buf, offset, TypeDBPointer, k)
// 	offset = writeString(om.buf, offset, v.Ref)
// 	copy(om.buf[offset:], v.ID[:])
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddCode ...
// func (om *OM) AddCode(k string, v Code) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + length + code + null
// 	// byte + optional scope document length
// 	dataSize := len(v.Code)
// 	if v.Scope != nil {
// 		dataSize += v.Scope.Len()
// 	}
// 	om.grow(7 + len(k) + dataSize)
// 	if v.Scope != nil {
// 		offset = writeTypeAndKey(om.buf, offset, TypeCodeWithScope, k)
// 	} else {
// 		offset = writeTypeAndKey(om.buf, offset, TypeJavaScript, k)
// 	}
// 	offset = writeString(om.buf, offset, v.Code)
// 	if v.Scope != nil {
// 		copy(om.buf[offset:], v.Scope.buf)
// 	}
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddSymbol ...
// func (om *OM) AddSymbol(k string, v string) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + string + null byte
// 	om.grow(7 + len(k) + len(v))
// 	offset = writeTypeAndKey(om.buf, offset, TypeSymbol, k)
// 	writeString(om.buf, offset, v)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddInt32 ...
// func (om *OM) AddInt32(k string, v int32) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + 4 for int32
// 	om.grow(6 + len(k))
// 	offset = writeTypeAndKey(om.buf, offset, TypeInt32, k)
// 	writeInt32(om.buf, offset, v)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddTimestamp ...
// func (om *OM) AddTimestamp(k string, v Timestamp) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + 8 typestamp bytes
// 	om.grow(10 + len(k))
// 	offset = writeTypeAndKey(om.buf, offset, TypeTimestamp, k)
// 	offset = writeUint32(om.buf, offset, v.Increment)
// 	writeUint32(om.buf, offset, v.Seconds)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddInt64 ...
// func (om *OM) AddInt64(k string, v int64) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + 8 for int64
// 	om.grow(10 + len(k))
// 	offset = writeTypeAndKey(om.buf, offset, TypeInt64, k)
// 	writeInt64(om.buf, offset, v)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddDecimal128 ...
// func (om *OM) AddDecimal128(k string, v Decimal128) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte + 16 for decimal128
// 	om.grow(18 + len(k))
// 	offset = writeTypeAndKey(om.buf, offset, TypeInt64, k)
// 	offset = writeUint64(om.buf, offset, v.L)
// 	writeUint64(om.buf, offset, v.H)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddMaxKey ...
// func (om *OM) AddMaxKey(k string) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte; no data bytes
// 	om.grow(2 + len(k))
// 	offset = writeTypeAndKey(om.buf, offset, TypeMaxKey, k)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
//
// // AddMinKey ...
// func (om *OM) AddMinKey(k string) *OM {
// 	if om.immutable || !om.valid {
// 		om.err = errImmutableInvalid
// 		return om
// 	}
// 	offset := len(om.buf) - 1
// 	// Add space for type byte + len(key) + null byte; no data bytes
// 	om.grow(2 + len(k))
// 	offset = writeTypeAndKey(om.buf, offset, TypeMinKey, k)
// 	om.buf[len(om.buf)-1] = 0
// 	return om
// }
