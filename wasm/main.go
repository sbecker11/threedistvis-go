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
