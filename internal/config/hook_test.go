package config_test

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"go.eloylp.dev/goomerang/internal/config"
	"go.eloylp.dev/goomerang/internal/test"
)

func TestOnStatusChangeHook(t *testing.T) {
	arbiter := test.NewArbiter(t)
	hooks := &config.Hooks{}
	hooks.AppendOnStatusChange(func(s uint32) {
		status := strconv.FormatUint(uint64(s), 10)
		arbiter.ItsAFactThat(fmt.Sprintf("HOOK_1_STATUS_%s_EXECUTED", status))
	})
	hooks.AppendOnStatusChange(func(s uint32) {
		status := strconv.FormatUint(uint64(s), 10)
		arbiter.ItsAFactThat(fmt.Sprintf("HOOK_2_STATUS_%s_EXECUTED", status))
	})
	hooks.ExecOnStatusChange(2)
	arbiter.RequireHappenedInOrder("HOOK_1_STATUS_2_EXECUTED", "HOOK_2_STATUS_2_EXECUTED")
}

func TestOnCloseHook(t *testing.T) {
	arbiter := test.NewArbiter(t)
	hooks := &config.Hooks{}
	hooks.AppendOnClose(func() {
		arbiter.ItsAFactThat("HOOK_1_EXECUTED")
	})
	hooks.AppendOnClose(func() {
		arbiter.ItsAFactThat("HOOK_2_EXECUTED")
	})
	hooks.ExecOnclose()
	arbiter.RequireHappenedInOrder("HOOK_1_EXECUTED", "HOOK_2_EXECUTED")
}

func TestOnErrorHook(t *testing.T) {
	arbiter := test.NewArbiter(t)
	hooks := &config.Hooks{}
	hooks.AppendOnError(func(err error) {
		arbiter.ItsAFactThat("HOOK_1_EXECUTED_" + err.Error())
	})
	hooks.AppendOnError(func(err error) {
		arbiter.ItsAFactThat("HOOK_2_EXECUTED_" + err.Error())
	})
	err := errors.New("ERROR")
	hooks.ExecOnError(err)
	arbiter.RequireHappenedInOrder("HOOK_1_EXECUTED_ERROR", "HOOK_2_EXECUTED_ERROR")
}
