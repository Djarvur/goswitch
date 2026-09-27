package engine

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/godbus/dbus/v5"
)

// Factory wire constants.
const (
	factoryPath   = dbus.ObjectPath("/org/freedesktop/IBus/Factory")
	ifaceFactory  = "org.freedesktop.IBus.Factory"
	enginePathFmt = "/org/freedesktop/IBus/Engine/goswitch/%d"
)

// factory mints per-input-context Engine objects for ibus-daemon.
type factory struct {
	conn    *dbus.Conn
	handler EventHandler
	mu      sync.Mutex
	seq     uint64
}

// CreateEngine implements org.freedesktop.IBus.Factory.CreateEngine
// (s engine_name) → (o object_path): it mints a monotonic object path,
// exports the engine object on all three interfaces and returns the path.
func (f *factory) CreateEngine(name string) (path dbus.ObjectPath, err *dbus.Error) {
	defer recoverHandler("CreateEngine", &err)

	f.mu.Lock()
	f.seq++
	n := f.seq
	f.mu.Unlock()

	path = dbus.ObjectPath(fmt.Sprintf(enginePathFmt, n))
	eng := NewEngine(f.handler, name)
	eng.conn = f.conn
	eng.path = path
	if f.handler != nil {
		// The handler (session.Actor) corrects through this engine's
		// emitters — the sink must be the bound object, so the attachment
		// happens exactly here, after conn and path are set (plan 02-03).
		f.handler.AttachEngine(eng)
	}
	for _, iface := range []string{ifaceEngine, ifaceService, ifaceProps} {
		if exportErr := f.conn.Export(eng, path, iface); exportErr != nil {
			slog.Error("engine export failed", "path", string(path), "error", exportErr)

			return "", &dbus.Error{
				Name: "org.freedesktop.DBus.Error.Failed",
				Body: []any{exportErr.Error()},
			}
		}
	}
	// The mode indicator registers at engine creation (owner decision 1,
	// quick plan 260927-way): the panel property with the daemon's start
	// symbol is emitted BEFORE the CreateEngine reply — every engine
	// registers its properties inside CreateEngine, and a fresh input
	// context must not start with a stale panel registration. Fire-and-
	// forget like every engine signal: a failed emit is error-logged, never
	// a registration failure.
	if err := f.conn.Emit(path, ifaceEngine+".RegisterProperties", dbus.MakeVariant(PropList{
		Name:        "IBusPropList",
		Attachments: map[string]dbus.Variant{},
		Properties:  []dbus.Variant{dbus.MakeVariant(NewModeProperty(initialModeSymbol))},
	})); err != nil {
		slog.Error("register properties emit failed", "path", string(path), "error", err)
	}
	slog.Info("engine created", "path", string(path), "name", name)

	return path, nil
}
