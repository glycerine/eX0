package main

import (
	"errors"
	"fmt"

	"github.com/shurcooL/go/gists/gist6545684"
	"github.com/shurcooL/gogl"
	glfw "github.com/shurcooL/goglfw"
)

func newLevel(name string) (*level, error) {
	l := new(level)

	err := l.initShaders()
	if err != nil {
		return nil, err
	}
	{
		f, err := glfw.Open(name)
		if err != nil {
			return nil, err
		}
		l.polygon, err = gist6545684.ReadGpcFromReader(f)
		f.Close()
		if err != nil {
			return nil, err
		}
	}
	err = l.createVbo()
	if err != nil {
		return nil, err
	}

	return l, nil
}

type level struct {
	polygon gist6545684.Polygon

	program                 *gogl.Program
	pMatrixUniform          *gogl.UniformLocation
	mvMatrixUniform         *gogl.UniformLocation
	vertexPositionBuffer    *gogl.Buffer
	vertexPositionAttribute int
}

const (
	levelVertexSource = `//#version 120 // OpenGL 2.1.
//#version 100 // WebGL.

attribute vec3 aVertexPosition;

uniform mat4 uMVMatrix;
uniform mat4 uPMatrix;

void main() {
	gl_Position = uPMatrix * uMVMatrix * vec4(aVertexPosition, 1.0);
}
`
	levelFragmentSource = `//#version 120 // OpenGL 2.1.
//#version 100 // WebGL.

void main() {
	gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0);
}
`
)

func (l *level) initShaders() error {
	vertexShader := gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(vertexShader, levelVertexSource)
	gl.CompileShader(vertexShader)
	defer gl.DeleteShader(vertexShader)

	if !gl.GetShaderParameterb(vertexShader, gl.COMPILE_STATUS) {
		return errors.New("COMPILE_STATUS: " + gl.GetShaderInfoLog(vertexShader))
	}

	fragmentShader := gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(fragmentShader, levelFragmentSource)
	gl.CompileShader(fragmentShader)
	defer gl.DeleteShader(fragmentShader)

	if !gl.GetShaderParameterb(fragmentShader, gl.COMPILE_STATUS) {
		return errors.New("COMPILE_STATUS: " + gl.GetShaderInfoLog(fragmentShader))
	}

	l.program = gl.CreateProgram()
	gl.AttachShader(l.program, vertexShader)
	gl.AttachShader(l.program, fragmentShader)

	gl.LinkProgram(l.program)
	if !gl.GetProgramParameterb(l.program, gl.LINK_STATUS) {
		return errors.New("LINK_STATUS: " + gl.GetProgramInfoLog(l.program))
	}

	gl.ValidateProgram(l.program)
	if !gl.GetProgramParameterb(l.program, gl.VALIDATE_STATUS) {
		return errors.New("VALIDATE_STATUS: " + gl.GetProgramInfoLog(l.program))
	}

	gl.UseProgram(l.program)

	l.pMatrixUniform = gl.GetUniformLocation(l.program, "uPMatrix")
	l.mvMatrixUniform = gl.GetUniformLocation(l.program, "uMVMatrix")

	if glError := gl.GetError(); glError != 0 {
		return fmt.Errorf("gl.GetError: %v", glError)
	}

	return nil
}

func (l *level) createVbo() error {
	l.vertexPositionBuffer = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vertexPositionBuffer)
	var vertices []float32
	for _, contour := range l.polygon.Contours {
		for _, vertex := range contour.Vertices {
			vertices = append(vertices, float32(vertex[0]), float32(vertex[1]))
		}
	}
	gl.BufferData(gl.ARRAY_BUFFER, vertices, gl.STATIC_DRAW)

	l.vertexPositionAttribute = gl.GetAttribLocation(l.program, "aVertexPosition")
	gl.EnableVertexAttribArray(l.vertexPositionAttribute)

	if glError := gl.GetError(); glError != 0 {
		return fmt.Errorf("gl.GetError: %v", glError)
	}

	return nil
}

func (l *level) setup() {
	gl.UseProgram(l.program)
	gl.BindBuffer(gl.ARRAY_BUFFER, l.vertexPositionBuffer)

	gl.VertexAttribPointer(l.vertexPositionAttribute, 2, gl.FLOAT, false, 0, 0)
}

func (l *level) render() {
	var first int
	for _, contour := range l.polygon.Contours {
		count := len(contour.Vertices)
		gl.DrawArrays(gl.LINE_LOOP, first, count)
		first += count
	}
}
