package pluginhost

import (
	"context"
	"reflect"
	"testing"
)

func TestExtensionPointExecutionModes(t *testing.T) {
	if !IsHookExtensionPoint(ExtensionPointAuthLoginSucceeded) {
		t.Fatalf(
			"expected %s to be hook extension point",
			ExtensionPointAuthLoginSucceeded,
		)
	}
	if !IsRegistrationExtensionPoint(ExtensionPointHTTPRouteRegister) {
		t.Fatalf(
			"expected %s to be registration extension point",
			ExtensionPointHTTPRouteRegister,
		)
	}
	if !IsExtensionPointExecutionModeSupported(
		ExtensionPointAuthLoginSucceeded,
		CallbackExecutionModeAsync,
	) {
		t.Fatalf(
			"expected %s to support %s mode",
			ExtensionPointAuthLoginSucceeded,
			CallbackExecutionModeAsync,
		)
	}
	if IsExtensionPointExecutionModeSupported(
		ExtensionPointHTTPRouteRegister,
		CallbackExecutionModeAsync,
	) {
		t.Fatalf(
			"expected %s to reject %s mode",
			ExtensionPointHTTPRouteRegister,
			CallbackExecutionModeAsync,
		)
	}
}

func TestCallbackInputContractsUseInterfaces(t *testing.T) {
	assertInterfaceType(t, (*HookPayload)(nil), "HookPayload")
	assertInterfaceType(t, (*AfterAuthInput)(nil), "AfterAuthInput")
	assertInterfaceType(t, (*RouteRegistrars)(nil), "RouteRegistrars")
	assertInterfaceType(t, (*CronRegistrar)(nil), "CronRegistrar")
	assertInterfaceType(t, (*MenuDescriptor)(nil), "MenuDescriptor")
	assertInterfaceType(t, (*PermissionDescriptor)(nil), "PermissionDescriptor")
}

func TestRegisterHookAcceptsAsyncMode(t *testing.T) {
	plugin := NewSourcePlugin("test-plugin-hook")
	plugin.RegisterHook(
		ExtensionPointAuthLoginSucceeded,
		CallbackExecutionModeAsync,
		func(ctx context.Context, payload HookPayload) error {
			return nil
		},
	)

	items := plugin.GetHookHandlers()
	if len(items) != 1 {
		t.Fatalf("expected one hook handler, got %d", len(items))
	}
	if items[0].Mode != CallbackExecutionModeAsync {
		t.Fatalf("expected async mode, got %s", items[0].Mode)
	}
}

func TestRegisterRoutesRejectsAsyncMode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected async route registration to panic")
		}
	}()

	plugin := NewSourcePlugin("test-plugin-route")
	plugin.RegisterRoutes(
		ExtensionPointHTTPRouteRegister,
		CallbackExecutionModeAsync,
		func(ctx context.Context, registrars RouteRegistrars) error {
			return nil
		},
	)
}

func TestCronRegistrarReportsPrimaryNode(t *testing.T) {
	registrar := NewCronRegistrar(
		"test-plugin",
		nil,
		func() bool { return false },
	)
	if registrar.IsPrimaryNode() {
		t.Fatalf("expected current node to be non-primary")
	}

	registrar = NewCronRegistrar(
		"test-plugin",
		nil,
		func() bool { return true },
	)
	if !registrar.IsPrimaryNode() {
		t.Fatalf("expected current node to be primary")
	}
}

func assertInterfaceType(t *testing.T, value interface{}, name string) {
	t.Helper()

	if reflect.TypeOf(value).Elem().Kind() != reflect.Interface {
		t.Fatalf("expected %s to be declared as interface", name)
	}
}
