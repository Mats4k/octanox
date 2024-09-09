package octanox

import "reflect"

type serializerRegistry map[reflect.Type]func(interface{}, Context) any

// Serializer is a type that represents a serializer function.
type Serializer func(interface{}, Context) any

// Serialize serializes an object into another form using the registered serializers.
func (i *Instance) Serialize(obj interface{}, c Context) any {
	serializer, ok := i.serializers[reflect.TypeOf(obj)]
	if !ok {
		return obj
	}
	return serializer(obj, c)
}

// RegisterSerializer is a function that registers a serializer for a given type.
func (i *Instance) RegisterSerializer(obj interface{}, serializer interface{}) *Instance {
	typeOfObj := reflect.TypeOf(obj)
	if _, ok := i.serializers[typeOfObj]; ok {
		panic("octanox: serializer for type " + typeOfObj.String() + " already registered")
	}

	ftype := reflect.ValueOf(serializer)
	i.serializers[typeOfObj] = func(obj interface{}, c Context) any {
		return ftype.Call([]reflect.Value{reflect.ValueOf(obj), reflect.ValueOf(c)})[0].Interface()
	}

	return i
}
