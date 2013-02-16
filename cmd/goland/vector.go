package main

import (
  "fmt"
  "math"
)

type Vector struct {
  X, Y float64
}

func (v *Vector) String() string {
  return fmt.Sprintf("[%.2f:%.2f]", v.X, v.Y)
}

func (v *Vector) Round() (int, int) {
  return int(v.X), int(v.Y)
}

func (v *Vector) Len() float64 {
  return math.Sqrt(float64(v.X*v.X + v.Y*v.Y))
}

func (v *Vector) Len2() float64 {
  return float64(v.X*v.X + v.Y*v.Y)
}

// subtract vector
func (v *Vector) Sub(o Vector) (* Vector) {
  v.X -= o.X
  v.Y -= o.Y
  return v
}

// add vector
func (v *Vector) Add(o Vector) (* Vector) {
  v.X += o.X
  v.Y += o.Y
  return v
}

// normalize vector
func (v *Vector) Normalize() (* Vector) {
  len := v.Len()
  if len != 0.0 {
    v.X /= len
    v.Y /= len
  }
  return v
}

// dot product
func (v *Vector) Dot(o Vector) float64 {
  return v.X * o.X + v.Y * o.Y
}

func (v *Vector) Scale(s float64) (* Vector) {
  v.X *= s
  v.Y *= s
  return v
}

func (v *Vector) Distance(o Vector) float64 {
  xd := o.X - v.X
  yd := o.Y - v.Y
  return math.Sqrt(xd*xd + yd*yd)
}

func (v *Vector) Distance2(o Vector) float64 {
  xd := o.X - v.X
  yd := o.Y - v.Y
  return xd*xd + yd*yd
}

func (v *Vector) Rotate(deg float64) (* Vector) {
  rads := deg * math.Pi/180
  cos := math.Cos(rads)
  sin := math.Sin(rads)

  v.X = v.X*cos - v.Y*sin
  v.Y = v.X*sin + v.Y*cos

  return v
}

// the angle in degrees of this vector (point) relative to the x-axis.
// Angles are counter-clockwise and between 0 and 360.
func (v *Vector) Angle() float64 {
  angle := math.Atan2(v.Y, v.X) * 180/math.Pi
  if angle < 0 {
    angle += 360
  }
  return angle
}

func (v *Vector) SetAngle(deg float64) {
  v = &Vector{v.Len(), 0.0}
  v.Rotate(deg)
}

func (v *Vector) Lerp(targ Vector, alpha float64) (* Vector) {
  return &Vector{}
}


// interpolation crap

type TwoDInterpolator struct {
  LastPos, LastVel Vector
}

type State struct {
  X, V float64
}

type Derivative struct {
  DX, DV float64
}

func interpolate(previous, current State, alpha float64) (s State) {
  s.X = current.X * alpha + previous.X * (1 - alpha)
  s.V = current.V * alpha + previous.V * (1 - alpha)
  return
}

func acceleration(state State, t float64) float64 {
  k := 10.0
  b := 1.0
  return - k*state.X - b*state.V
}

func evaluate1(initial State, t float64) (der Derivative) {
  der.DX = initial.V
  der.DV = acceleration(initial, t)
  return
}

func evaluate2(initial State, t, dt float64, d Derivative) (der Derivative) {
  s := State{X: initial.X + d.DX * dt, V: initial.V + d.DV * dt}

  der.DX = s.V
  der.DV = acceleration(s, t+dt)
  return
}

func integrate(state State, t, dt float64) (s State) {
  a := evaluate1(state, t)
  b := evaluate2(state, t, dt * 0.5, a)
  c := evaluate2(state, t, dt * 0.5, b)
  d := evaluate2(state, t, dt, c)

  dxdt := 1.0/6.0 * (a.DX + 2.0 * (b.DX + c.DX) + d.DX)
  dvdt := 1.0/6.0 * (a.DV + 2.0 * (b.DV + c.DV) + d.DV)

  s.X = state.X + dxdt * dt
  s.V = state.V + dvdt * dt
  return
}

