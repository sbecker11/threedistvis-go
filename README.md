# ThreeDistVis-Go

A GoLang-based version of ThreeDistVis, a 3D visualization tool for rendering point distributions in a web browser. Uses Go for the backend and WebAssembly for client-side WebGL rendering.

## Prerequisites

- **GoLang**: Version 1.21 or later.
- **macOS**: Compatible with your 2019 MacBook Pro (macOS Ventura or later recommended).
- **Browser**: Chrome, Firefox, or Safari for WebAssembly support.
- **Git**: For version control and GitHub integration.

## Project Structure

```
threedistvis-go/
├── main.go                # Go backend server
├── wasm/                  # WebAssembly frontend
│   ├── main.go            # WASM entry point for 3D rendering
│   ├── index.html         # HTML template
│   ├── wasm_exec.js       # Go WASM runtime
│   └── styles.css         # Basic styling
├── go.mod                 # Go module definition
├── README.md              # Project documentation
└── .gitignore             # Git ignore file
```

## Setup Instructions

1. **Install GoLang**:
   - Download and install Go from https://go.dev/dl/ (version 1.21 or later).
   - Verify installation: `go version` (should output `go version go1.21.x darwin/amd64`).
   - Set up your GOPATH: `export GOPATH=$HOME/go` in `~/.zshrc` or `~/.bash_profile`.

2. **Clone the Repository**:
   ```bash
   git clone https://github.com/sbecker11/threedistvis-go.git
   cd threedistvis-go
   ```

3. **Initialize Go Module** (if not already done):
   ```bash
   go mod init github.com/sbecker11/threedistvis-go
   ```

4. **Install Dependencies**:
   - No external Go packages are required for this project (uses standard library).

5. **Build and Run Locally**:
   - Build the WASM frontend:
     ```bash
     cd wasm
     GOOS=js GOARCH=wasm go build -o main.wasm main.go
     cd ..
     ```
   - Run the Go backend:
     ```bash
     go run main.go
     ```
   - Open `http://localhost:8080` in your browser to view the 3D visualization.

6. **Deploy to GitHub Pages**:
   - Copy `wasm/index.html`, `wasm/main.wasm`, `wasm/wasm_exec.js`, and `wasm/styles.css` to the root of your GitHub repository’s `main` branch.
   - Enable GitHub Pages in the repository settings, pointing to the `main` branch.
   - Access at `https://sbecker11.github.io/threedistvis-go/`.

## Backend Code (main.go)

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Serve static files from the wasm directory
	fs := http.FileServer(http.Dir("wasm"))
	http.Handle("/", fs)

	fmt.Println("Server running at http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Server error:", err)
	}
}
```

## WASM Frontend Code (wasm/main.go)

```go
package main

import (
	"math/rand"
	"syscall/js"
)

func main() {
	// Ensure the WASM module stays alive
	c := make(chan struct{}, 0)

	// Get canvas and WebGL context
	canvas := js.Global().Get("document").Call("getElementById", "canvas")
	gl := canvas.Call("getContext", "webgl")
	if gl.IsNull() {
		js.Global().Call("alert", "WebGL not supported")
		return
	}

	// Set up WebGL
	gl.Call("clearColor", 0.0, 0.0, 0.0, 1.0)
	gl.Call("clear", gl.Get("COLOR_BUFFER_BIT"))

	// Create simple 3D points (random example)
	points := make([]float32, 300) // 100 points (x, y, z)
	for i := 0; i < 300; i++ {
		points[i] = rand.Float32()*2 - 1 // Random between -1 and 1
	}

	// Set up vertex buffer
	buffer := gl.Call("createBuffer")
	gl.Call("bindBuffer", gl.Get("ARRAY_BUFFER"), buffer)
	gl.Call("bufferData", gl.Get("ARRAY_BUFFER"), js.TypedArrayOf(points), gl.Get("STATIC_DRAW"))

	// Simple vertex shader
	vertexShaderSource := `
		attribute vec3 position;
		uniform mat4 modelViewProjection;
		void main() {
			gl_Position = modelViewProjection * vec4(position, 1.0);
			gl_PointSize = 5.0;
		}
	`
	vertexShader := gl.Call("createShader", gl.Get("VERTEX_SHADER"))
	gl.Call("shaderSource", vertexShader, vertexShaderSource)
	gl.Call("compileShader", vertexShader)

	// Simple fragment shader
	fragmentShaderSource := `
		precision mediump float;
		void main() {
			gl_FragColor = vec4(1.0, 1.0, 1.0, 1.0); // White points
		}
	`
	fragmentShader := gl.Call("createShader", gl.Get("FRAGMENT_SHADER"))
	gl.Call("shaderSource", fragmentShader, fragmentShaderSource)
	gl.Call("compileShader", fragmentShader)

	// Create shader program
	program := gl.Call("createProgram")
	gl.Call("attachShader", program, vertexShader)
	gl.Call("attachShader", program, fragmentShader)
	gl.Call("linkProgram", program)
	gl.Call("useProgram", program)

	// Set up attribute and uniform
	positionLoc := gl.Call("getAttribLocation", program, "position")
	gl.Call("enableVertexAttribArray", positionLoc)
	gl.Call("vertexAttribPointer", positionLoc, 3, gl.Get("FLOAT"), false, 0, 0)

	mvpLoc := gl.Call("getUniformLocation", program, "modelViewProjection")

	// Simple animation loop
	var angle float32
	render := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		angle += 0.01

		// Simple rotation matrix (for demo)
		s, c := float32(math.Sin(float64(angle))), float32(math.Cos(float64(angle)))
		mvp := []float32{
			c, 0, s, 0,
			0, 1, 0, 0,
			-s, 0, c, 0,
			0, 0, -2, 1,
		}

		gl.Call("uniformMatrix4fv", mvpLoc, false, js.TypedArrayOf(mvp))
		gl.Call("clear", gl.Get("COLOR_BUFFER_BIT"))
		gl.Call("drawArrays", gl.Get("POINTS"), 0, 100)

		js.Global().Call("requestAnimationFrame", render)
		return nil
	})
	js.Global().Call("requestAnimationFrame", render)

	<-c
}
```

## HTML Template (wasm/index.html)

```html
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>ThreeDistVis-Go</title>
	<link rel="stylesheet" href="styles.css">
</head>
<body>
	<canvas id="canvas" width="800" height="600"></canvas>
	<script src="wasm_exec.js"></script>
	<script>
		const go = new Go();
		WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
			go.run(result.instance);
		});
	</script>
</body>
</html>
```

## WASM Runtime (wasm/wasm_exec.js)

- Copy this file from `$GOROOT/misc/wasm/wasm_exec.js` (included with Go installation).
- Alternatively, run:
  ```bash
  cp $(go env GOROOT)/misc/wasm/wasm_exec.js wasm/
  ```

## Styling (wasm/styles.css)

```css
body {
	margin: 0;
	display: flex;
	justify-content: center;
	align-items: center;
	height: 100vh;
	background-color: #222;
}

canvas {
	border: 1px solid #444;
}
```

## .gitignore

```
# Binaries and build outputs
*.wasm
*.exe
*.o
*.a

# Dependency directories
/vendor/

# Editor and IDE files
.vscode/
.idea/
*.swp
*.swo
```

## Notes

- **Functionality**: Renders 100 random 3D points with rotation animation using WebGL via WebAssembly. You can extend this by adding controls (e.g., mouse-based rotation, zoom) or loading specific point data.
- **Performance**: Optimized for your 2019 MacBook Pro. WebAssembly ensures smooth rendering in modern browsers.
- **Deployment**: Static files (`index.html`, `main.wasm`, `wasm_exec.js`, `styles.css`) are ready for GitHub Pages.
- **Extensibility**: Add API endpoints in `main.go` to serve dynamic data or integrate with your original project’s data sources.

## Troubleshooting

- **WASM Build Fails**: Ensure `GOOS=js` and `GOARCH=wasm` are set. Verify Go version.
- **Port Conflict**: If `8080` is in use, check with `lsof -i :8080` and kill processes with `kill -9 <pid>` (from our May 13, 2025 discussion).
- **Browser Issues**: Use Chrome for best WebAssembly support. Clear cache if rendering fails.