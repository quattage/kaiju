/******************************************************************************/
/* id.go                                                                      */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package klib

import (
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
)

var allowed = regexp.MustCompile(`[^a-z_.]`)

type RegistryKey struct {
	name string
	id   string
}

type RegistryProvider interface {
	GetRegistryKey() RegistryKey
}

func NewKey(id, name string) RegistryKey {
	name = strings.TrimSpace(name)
	if len(name) <= 0 {
		name = "unnamed"
	}
	id = strings.TrimSpace(id)
	id = allowed.ReplaceAllString(strings.ToLower(id), "")
	if len(id) <= 0 {
		panic("can't create registry key with empty ID (from key named '" + name + "')")
	}
	output := RegistryKey{name, id}
	return output
}

func (id RegistryKey) Name() string {
	return id.name
}

func (id RegistryKey) ID() string {
	return id.id
}

func (id RegistryKey) String() string {
	return id.id + " (" + id.name + ")"
}

func (id RegistryKey) Equals(other RegistryKey) bool {
	return id.id == other.id
}

func (id RegistryKey) IsNamed(name string) bool {
	return strings.EqualFold(id.name, name)
}

// A Registry is a type-specific map for self-identifying registry objects.
// Registry objects can be pulled from the backing map via string or
// RegistryKey. Duplicate elements will not overwrite. Mappings to nil pointers
// will be ignored. Entries cannot be removed once added.
type Registry[T RegistryProvider] struct {
	key      RegistryKey
	contents map[string]T
}

func NewRegistry[T RegistryProvider](id, name string) Registry[T] {
	return Registry[T]{
		key:      NewKey(id, name),
		contents: make(map[string]T, 32),
	}
}

func (r *Registry[T]) GetRegistryKey() RegistryKey {
	return r.key
}

func (r *Registry[T]) AddUnique(obj T) {
	if r.contents == nil {
		r.contents = make(map[string]T, 32)
	}
	vt := reflect.ValueOf(obj)
	if vt.IsValid() && vt.Kind() == reflect.Pointer && vt.IsNil() {
		slog.Error("Skipped adding a nil mapping to a registry")
	}
	rkey := obj.GetRegistryKey()
	_, exists := r.contents[rkey.id]
	if exists {
		slog.Error(fmt.Sprintf("Couldn't register an entry id ID '%s' - A mapping at this ID already exists!", rkey.id))
		return
	}
	r.contents[rkey.id] = obj
}

func (r *Registry[T]) GetByString(id string) (T, bool) {
	if r.contents == nil {
		var zero T
		return zero, false
	}
	output, ok := r.contents[id]
	return output, ok
}

func (r *Registry[T]) GetByKey(key RegistryKey) (T, bool) {
	if r.contents == nil {
		var zero T
		return zero, false
	}
	output, ok := r.contents[key.id]
	return output, ok
}

func (r *Registry[T]) GetByStringOrPanic(id string) T {
	output, ok := r.contents[id]
	if !ok {
		panic(fmt.Sprintf("no key '%v' in registry", id))
	}
	return output
}

func (r *Registry[T]) GetByKeyOrPanic(key RegistryKey) T {
	output, ok := r.GetByKey(key)
	if !ok {
		panic(fmt.Sprintf("no key '%v' in registry", key))
	}
	return output
}

func (r *Registry[T]) IsEmpty() bool {
	return len(r.contents) <= 0
}

func (r *Registry[T]) FormatContents() string {
	sb := strings.Builder{}
	sb.WriteString(r.key.ID())
	sb.WriteString("[\r\n")
	if len(r.contents) <= 0 {
		sb.WriteString("empty\n")
		sb.WriteString("]")
		return sb.String()
	}
	for key := range r.contents {
		sb.WriteString("  - ")
		sb.WriteString(key)
		sb.WriteString("\n")
	}
	sb.WriteString("]")
	return sb.String()
}
