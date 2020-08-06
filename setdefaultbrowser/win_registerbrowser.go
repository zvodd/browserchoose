package setdefaultbrowser

// Approuch is to set the app as the browser, not the actually selected browser.
// Then use a proxy launch / syscall to actually launch the browser.
// This means users can redirect url handlers to things like wget or shell scripts
// or specify a profile with arguments to the browser executable.
// or have a window pop app to select the browser option per url call.
