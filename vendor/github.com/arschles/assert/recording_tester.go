package assert

type fatalRecording struct {
	str  string
	args []interface{}
}

func newFatalRecording(str string, args []interface{}) *fatalRecording {
	return &fatalRecording{str: str, args: args}
}

type recordingTester struct {
	fatals []fatalRecording
}

func newRecordingTester() *recordingTester {
	return &recordingTester{fatals: nil}
}

func (r *recordingTester) fatalsLen() int {
	return len(r.fatals)
}

func (r *recordingTester) Fatalf(fmt string, args ...interface{}) {
	r.fatals = append(r.fatals, *newFatalRecording(fmt, args))
}
