package sigmon

import (
	"reflect"
	"testing"
)

func TestHandlerFuncRegistry(t *testing.T) {
	t.Run("LoadBufferBuffer", tHandlerFuncRegistryLoadBufferBuffer)
	t.Run("SetGet", tHandlerFuncRegistrySetGet)
}

func tHandlerFuncRegistryLoadBufferBuffer(t *testing.T) {
	r := newHandlerFuncRegistry(nil)
	f := func(*State) {}

	r.loadBuffer(nil)
	r.loadBuffer(f)

	select {
	case fn := <-r.buffer():
		got := reflect.ValueOf(fn).Pointer()
		want := reflect.ValueOf(f).Pointer()
		if got != want {
			t.Fatalf("got %v, want %v", got, want)
		}

	default:
		t.Fatalf("should not block")
	}

	r.loadBuffer(nil)

	select {
	case fn := <-r.buffer():
		got := reflect.ValueOf(fn).Pointer()
		notWant := reflect.ValueOf(f).Pointer()
		if got == notWant {
			t.Fatalf("handlerfunc not replaced")
		}
		if got == 0x0 {
			t.Fatalf("handlerfunc should not be nil")
		}

	default:
		t.Fatalf("should not block")
	}
}

func tHandlerFuncRegistrySetGet(t *testing.T) {
	r := newHandlerFuncRegistry(nil)
	f := func(*State) {}

	if r.get() == nil {
		t.Fatalf("got nil, want empty func")
	}

	r.set(nil)
	r.set(f)

	got := reflect.ValueOf(r.get()).Pointer()
	want := reflect.ValueOf(f).Pointer()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	r.set(nil)

	got = reflect.ValueOf(r.get()).Pointer()
	notWant := reflect.ValueOf(f).Pointer()
	if got == notWant {
		t.Fatalf("handlerfunc not replaced")
	}
	if got == 0x0 {
		t.Fatalf("handlerfunc should not be nil")
	}
}
