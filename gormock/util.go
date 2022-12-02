package e2e

type TestingTB interface {
	// Name Returns current test name.
	Name() string
	Cleanup(f func())
	Logf(fmt string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Errorf(message string, args ...interface{})
}
