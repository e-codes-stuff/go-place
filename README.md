# go-place, an r/place inspired image API

This program provides an API (created with gin) which allows manipulation of pixels on a board and viewing of the boards current state.

Each pixel placed is stored in a database using gorm, which means the entire board history is preserved.
