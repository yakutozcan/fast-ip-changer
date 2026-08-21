package main

import (
	"context"

	"github.com/yakutozcan/fast-ip-changer/pkg/diagnostics"
	"github.com/yakutozcan/fast-ip-changer/pkg/network"
	"github.com/yakutozcan/fast-ip-changer/pkg/profile"
	"github.com/yakutozcan/fast-ip-changer/pkg/sysexec"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// context returns the app context, falling back to a background context when a
// bound method is somehow called before startup ran.
func (a *App) context() context.Context {
	if a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

// Network Methods

func (a *App) GetAdapters() ([]network.Adapter, error) {
	return network.GetAdapters(a.context())
}

func (a *App) SetStaticIP(adapterName, ip, subnet, gateway, dns string) error {
	return network.SetStaticIP(a.context(), adapterName, ip, subnet, gateway, dns)
}

func (a *App) SetDHCP(adapterName string) error {
	return network.SetDHCP(a.context(), adapterName)
}

func (a *App) EnableAdapter(adapterName string) error {
	return network.EnableAdapter(a.context(), adapterName)
}

func (a *App) DisableAdapter(adapterName string) error {
	return network.DisableAdapter(a.context(), adapterName)
}

// IsElevated reports whether the app can change the network configuration
// without further prompting, so the UI can warn the user up front.
func (a *App) IsElevated() bool {
	return sysexec.IsElevated()
}

// Profile Methods

func (a *App) GetProfiles() ([]profile.IPProfile, error) {
	return profile.GetProfiles()
}

func (a *App) AddProfile(p profile.IPProfile) error {
	return profile.AddProfile(p)
}

func (a *App) UpdateProfile(p profile.IPProfile) error {
	return profile.UpdateProfile(p)
}

func (a *App) DeleteProfile(id string) error {
	return profile.DeleteProfile(id)
}

func (a *App) OpenProfileFolder() error {
	return profile.OpenProfileFolder()
}

// Diagnostics Methods

func (a *App) PingHost(host string, count int) (string, error) {
	return diagnostics.PingHost(a.context(), host, count)
}

func (a *App) TraceRouteHost(host string) (string, error) {
	return diagnostics.TraceRouteHost(a.context(), host)
}

// CancelDiagnostics stops a running ping or traceroute.
func (a *App) CancelDiagnostics() {
	diagnostics.Cancel()
}

func (a *App) QuickCheck(gateway string, checkPublicIP bool) diagnostics.QuickCheckResult {
	return diagnostics.QuickCheck(a.context(), gateway, checkPublicIP)
}
