package ansiterm

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func getStateNames() []string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	parser, _ := createTestParser(ctx, "Ground")

	stateNames := []string{}
	for _, state := range parser.stateMap {
		stateNames = append(stateNames, state.Name())
	}

	return stateNames
}

func stateTransitionHelper(t *testing.T, start string, end string, bytes []byte) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, b := range bytes {
		bytes := []byte{byte(b)}
		parser, _ := createTestParser(ctx, start)
		parser.Parse(bytes)
		validateState(t, start, parser.currState, end)
	}
}

func anyToXHelper(t *testing.T, bytes []byte, expectedState string) {
	for _, s := range getStateNames() {
		stateTransitionHelper(t, s, expectedState, bytes)
	}
}

func funcCallParamHelper(t *testing.T, bytes []byte, start string, expected string, expectedCalls []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	parser, evtHandler := createTestParser(ctx, start)
	parser.Parse(bytes)
	validateState(t, start, parser.currState, expected)

	err := evtHandler.waitForNCalls(1, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	validateFuncCalls(t, evtHandler.FunctionCalls, expectedCalls)
}

func parseParamsHelper(t *testing.T, bytes []byte, expectedParams []string) {
	params, err := parseParams(bytes)

	if err != nil {
		t.Errorf("Parameter parse error: %v", err)
		return
	}

	if len(params) != len(expectedParams) {
		t.Errorf("Parsed   parameters: %v", params)
		t.Errorf("Expected parameters: %v", expectedParams)
		t.Errorf("Parameter length failure: %d != %d", len(params), len(expectedParams))
		return
	}

	for i, v := range expectedParams {
		if v != params[i] {
			t.Errorf("Parsed   parameters: %v", params)
			t.Errorf("Expected parameters: %v", expectedParams)
			t.Errorf("Parameter parse failure: %s != %s at position %d", v, params[i], i)
		}
	}
}

func cursorSingleParamHelper(t *testing.T, command byte, funcName string) {
	funcCallParamHelper(t, []byte{command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([1])", funcName)})
	funcCallParamHelper(t, []byte{'0', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([1])", funcName)})
	funcCallParamHelper(t, []byte{'2', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([2])", funcName)})
	funcCallParamHelper(t, []byte{'2', '3', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([23])", funcName)})
	funcCallParamHelper(t, []byte{'2', ';', '3', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([2])", funcName)})
	funcCallParamHelper(t, []byte{'2', ';', '3', ';', '4', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([2])", funcName)})
}

func cursorTwoParamHelper(t *testing.T, command byte, funcName string) {
	funcCallParamHelper(t, []byte{command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([1 1])", funcName)})
	funcCallParamHelper(t, []byte{'0', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([1 1])", funcName)})
	funcCallParamHelper(t, []byte{'2', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([2 1])", funcName)})
	funcCallParamHelper(t, []byte{'2', '3', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([23 1])", funcName)})
	funcCallParamHelper(t, []byte{'2', ';', '3', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([2 3])", funcName)})
	funcCallParamHelper(t, []byte{'2', ';', '3', ';', '4', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([2 3])", funcName)})
}

func eraseHelper(t *testing.T, command byte, funcName string) {
	funcCallParamHelper(t, []byte{command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([0])", funcName)})
	funcCallParamHelper(t, []byte{'0', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([0])", funcName)})
	funcCallParamHelper(t, []byte{'1', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([1])", funcName)})
	funcCallParamHelper(t, []byte{'2', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([2])", funcName)})
	funcCallParamHelper(t, []byte{'3', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([3])", funcName)})
	funcCallParamHelper(t, []byte{'4', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([0])", funcName)})
	funcCallParamHelper(t, []byte{'1', ';', '2', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([1])", funcName)})
}

func scrollHelper(t *testing.T, command byte, funcName string) {
	funcCallParamHelper(t, []byte{command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([1])", funcName)})
	funcCallParamHelper(t, []byte{'0', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([1])", funcName)})
	funcCallParamHelper(t, []byte{'1', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([1])", funcName)})
	funcCallParamHelper(t, []byte{'5', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([5])", funcName)})
	funcCallParamHelper(t, []byte{'4', ';', '6', command}, "CsiEntry", "Ground", []string{fmt.Sprintf("%s([4])", funcName)})
}

func clearOnStateChangeHelper(t *testing.T, start string, end string, bytes []byte) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p, _ := createTestParser(ctx, start)
	fillContext(p.context)
	p.Parse(bytes)
	validateState(t, start, p.currState, end)
	validateEmptyContext(t, p.context)
}

func c0Helper(t *testing.T, bytes []byte, expectedState string, expectedCalls []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	parser, evtHandler := createTestParser(ctx, "Ground")
	parser.Parse(bytes)
	validateState(t, "Ground", parser.currState, expectedState)
	validateFuncCalls(t, evtHandler.FunctionCalls, expectedCalls)
}
