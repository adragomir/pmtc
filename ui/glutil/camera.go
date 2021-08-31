package glutil

import (
	"log"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Object struct {
	position  mgl32.Vec3
	scale     mgl32.Vec3
	direction mgl32.Vec3
	rotation  mgl32.Vec3
	quat      mgl32.Quat

	// Lccal transform matrix stores position/rotation/scale relative to parent
	matrix mgl32.Mat4
	// World transform matrix stores position/rotation/scale relative to highest ancestor (generally the scene)
	matrixWorld mgl32.Mat4

	matNeedsUpdate bool
	rotNeedsUpdate bool
}

func (o *Object) Init() {
	o.position = mgl32.Vec3{0, 0, 0}
	o.scale = mgl32.Vec3{1, 1, 1}
	o.direction = mgl32.Vec3{0, 0, 1}
	o.rotation = mgl32.Vec3{0, 0, 0}
	o.quat = mgl32.QuatIdent()
	o.matrix = mgl32.Ident4()
	o.matrixWorld = mgl32.Ident4()
}
func (o *Object) SetChanged(changed bool) {
	o.matNeedsUpdate = changed
}
func (o *Object) SetPosition(pos mgl32.Vec3) {
	o.position = pos
	o.matNeedsUpdate = true
}
func (o *Object) TranslateOnAxis(axis mgl32.Vec3, dist float32) {
	v := mgl32.Vec3{axis[0], axis[1], axis[2]}
	v = o.quat.Rotate(v)
	v = v.Mul(dist)
	o.position = o.position.Add(v)
	o.matNeedsUpdate = true
}

func (o *Object) Position() mgl32.Vec3 {
	return o.position
}

func (o *Object) SetRotation(r mgl32.Vec3) {
	o.rotation = r
	o.quat = SetFromEuler(o.rotation)
	o.matNeedsUpdate = true
}

func (o *Object) SetRotationQuat(q mgl32.Quat) {
	o.quat = q
	o.rotNeedsUpdate = true
}

func (o *Object) SetRotationX(x float32) {
	if o.rotNeedsUpdate {
		o.rotation = SetVectorFromQuat(o.quat)
		o.rotNeedsUpdate = false
	}
	o.rotation[0] = x
	o.quat = SetFromEuler(o.rotation)
	o.matNeedsUpdate = true
}

func (o *Object) SetRotationY(y float32) {
	if o.rotNeedsUpdate {
		o.rotation = SetVectorFromQuat(o.quat)
		o.rotNeedsUpdate = false
	}
	o.rotation[1] = y
	o.quat = SetFromEuler(o.rotation)
	o.matNeedsUpdate = true
}

// SetRotationZ sets the global Z rotation to the specified angle in radians.
func (o *Object) SetRotationZ(z float32) {
	if o.rotNeedsUpdate {
		o.rotation = SetVectorFromQuat(o.quat)
		o.rotNeedsUpdate = false
	}
	o.rotation[2] = z
	o.quat = SetFromEuler(o.rotation)
	o.matNeedsUpdate = true
}
func (o *Object) Rotation() mgl32.Vec3 {
	if o.rotNeedsUpdate {
		o.rotation = SetVectorFromQuat(o.quat)
		o.rotNeedsUpdate = false
	}
	return o.rotation
}

func (o *Object) RotateOnAxis(axis mgl32.Vec3, angle float32) {
	q := mgl32.QuatRotate(angle, axis)
	o.QuaternionMult(q)
}

func (o *Object) RotateX(x float32) {
	o.RotateOnAxis(mgl32.Vec3{1, 0, 0}, x)
}

func (o *Object) RotateY(y float32) {
	o.RotateOnAxis(mgl32.Vec3{0, 1, 0}, y)
}

func (o *Object) RotateZ(z float32) {
	o.RotateOnAxis(mgl32.Vec3{0, 0, 1}, z)
}

func (o *Object) SetQuaternion(q mgl32.Quat) {
	o.quat = q
	o.rotNeedsUpdate = true
}

func (o *Object) QuaternionMult(q mgl32.Quat) {
	o.quat = o.quat.Mul(q)
	o.rotNeedsUpdate = true
}

func (o *Object) WorldPosition() mgl32.Vec3 {
	o.UpdateMatrixWorld()
	return mgl32.Vec3{o.matrixWorld[12], o.matrixWorld[13], o.matrixWorld[14]}
}

// LookAt rotates the node to look at the specified target position, using the specified up vector.
func (o *Object) LookAt(target, up mgl32.Vec3) {
	worldPos := o.WorldPosition()
	rotMat := LookAt(worldPos, target, up)
	o.quat = mgl32.Mat4ToQuat(rotMat)
	o.rotNeedsUpdate = true
}
func (o *Object) SetDirection(v mgl32.Vec3) {
	o.direction = v
	o.matNeedsUpdate = true
}
func (o *Object) SetMatrix(m mgl32.Mat4) {
	o.matrix = m
	o.position, o.quat, o.scale = DecomposeMatrixToPosQuatScale(o.matrix)
	o.rotNeedsUpdate = true
}

func (o *Object) WorldQuaternion() mgl32.Quat {
	o.UpdateMatrixWorld()
	pos, q, scale := DecomposeMatrixToPosQuatScale(o.matrixWorld)
	o.position = pos
	o.scale = scale
	return q
}
func (o *Object) WorldRotation() mgl32.Vec3 {
	q := o.WorldQuaternion()
	return SetVectorFromQuat(q)
}

// WorldScale updates the world matrix and sets
// the specified vector to the current world scale of this node.
func (o *Object) WorldScale() mgl32.Vec3 {
	o.UpdateMatrixWorld()
	_, _, s := DecomposeMatrixToPosQuatScale(o.matrixWorld)
	return s
}

// WorldDirection updates the world matrix and sets
// the specified vector to the current world direction of this node.
func (o *Object) WorldDirection() mgl32.Vec3 {
	q := o.WorldQuaternion()
	//return ApplyQuat(q, o.direction)
	return q.Rotate(o.direction)
}

func (o *Object) MatrixWorld() mgl32.Mat4 {
	return o.matrixWorld
}

func (o *Object) UpdateMatrix() bool {
	if !o.matNeedsUpdate && !o.rotNeedsUpdate {
		return false
	}
	o.matrix = Compose(o.position, o.quat, o.scale)
	o.matNeedsUpdate = false
	return true
}

func (o *Object) UpdateMatrixWorld() {
	o.UpdateMatrix()
	o.matrixWorld = o.matrix
}

type ProjectionType int

// The possible camera projections.
const (
	Perspective = ProjectionType(iota)
	Orthographic
)

type Camera struct {
	Object

	fov    float32
	aspect float32
	near   float32
	far    float32

	size float32
	proj ProjectionType

	projChanged bool
	projMatrix  mgl32.Mat4
	viewMatrix  mgl32.Mat4
}

func NewPerspective(fov, aspect, near, far float32) *Camera {
	c := new(Camera)
	c.Object.Init()
	c.fov = fov
	c.aspect = aspect
	c.near = near
	c.far = far
	c.proj = Perspective
	c.size = 8
	c.projChanged = true
	return c
}

func NewOrthographic(aspect, near, far, size float32) *Camera {
	c := new(Camera)
	c.Object.Init()
	c.fov = 60
	c.aspect = aspect
	c.near = near
	c.far = far

	c.proj = Orthographic
	c.size = size

	c.projChanged = true
	return c
}

func (c *Camera) SetAspect(aspect float32) {
	if aspect == c.aspect {
		return
	}
	c.aspect = aspect
	c.projChanged = true
}

func (c *Camera) UpdateFov(targetDist float32) {
	c.fov = float32(2 * mgl32.Atan(c.size/(2*targetDist)) * 180 / math.Pi)
	if c.proj == Perspective {
		c.projChanged = true
	}
}

func (c *Camera) UpdateSize(targetDist float32) {
	c.size = 2 * targetDist * float32(mgl32.Tan(mgl32.DegToRad(c.fov*0.5)))
	if c.proj == Orthographic {
		c.projChanged = true
	}
}

// ViewMatrix returns the view matrix of the camera.
func (c *Camera) ViewMatrix() *mgl32.Mat4 {
	c.UpdateMatrixWorld()
	matrixWorld := c.MatrixWorld()
	c.viewMatrix = matrixWorld.Inv()
	return &c.viewMatrix
}

func (c *Camera) ViewMatrixGl() *float32 {
	return &c.ViewMatrix()[0]
}

// ProjMatrix returns the projection matrix of the camera.
func (c *Camera) ProjMatrix() *mgl32.Mat4 {
	if c.projChanged {
		switch c.proj {
		case Perspective:
			t := c.near * mgl32.Tan(mgl32.DegToRad(c.fov*0.5))
			ymax := t
			ymin := -t
			xmax := t
			xmin := -t
			ymax /= c.aspect
			ymin /= c.aspect
			c.projMatrix = mgl32.Frustum(xmin, xmax, ymin, ymax, c.near, c.far)
		case Orthographic:
			s := c.size / 2
			var h, w float32
			h = s / c.aspect
			w = s
			c.projMatrix = mgl32.Ortho(-w, w, h, -h, c.near, c.far)
		}
		c.projChanged = false
	}
	return &c.projMatrix
}

func (c *Camera) ProjMatrixGl() *float32 {
	return &c.ProjMatrix()[0]
}

// OrbitEnabled specifies which control types are enabled.
type OrbitEnabled int

// The possible control types.
const (
	OrbitNone OrbitEnabled = 0x00
	OrbitRot  OrbitEnabled = 0x01
	OrbitZoom OrbitEnabled = 0x02
	OrbitPan  OrbitEnabled = 0x04
	OrbitKeys OrbitEnabled = 0x08
	OrbitAll  OrbitEnabled = 0xFF
)

// orbitState bitmask
type orbitState int

const (
	stateNone = orbitState(iota)
	stateRotate
	stateZoom
	statePan
)

type CameraControl struct {
	camera  *Camera
	target  mgl32.Vec3
	up      mgl32.Vec3
	enabled OrbitEnabled // Which controls are enabled
	state   orbitState   // Current control state

	MinDistance     float32 // Minimum distance from target (default is 1)
	MaxDistance     float32 // Maximum distance from target (default is infinity)
	MinPolarAngle   float32 // Minimum polar angle in radians (default is 0)
	MaxPolarAngle   float32 // Maximum polar angle in radians (default is Pi)
	MinAzimuthAngle float32 // Minimum azimuthal angle in radians (default is negative infinity)
	MaxAzimuthAngle float32 // Maximum azimuthal angle in radians (default is infinity)
	RotSpeed        float32 // Rotation speed factor (default is 1)
	ZoomSpeed       float32 // Zoom speed factor (default is 0.1)
	KeyRotSpeed     float32 // Rotation delta in radians used on each rotation key event (default is the equivalent of 15 degrees)
	KeyZoomSpeed    float32 // Zoom delta used on each zoom key event (default is 2)
	KeyPanSpeed     float32 // Pan delta used on each pan key event (default is 35)

	// Internal
	mousePos  mgl32.Vec2
	rotStart  mgl32.Vec2
	panStart  mgl32.Vec2
	zoomStart float32
	areaSize  [2]float32
}

func NewCameraControl(c *Camera) *CameraControl {
	cc := new(CameraControl)
	cc.camera = c
	cc.target = mgl32.Vec3{0, 0, 0}
	cc.up = mgl32.Vec3{0, 1, 0}
	cc.enabled = OrbitAll

	cc.MinDistance = 1.0
	cc.MaxDistance = float32(mgl32.Inf(1))
	cc.MinPolarAngle = 0
	cc.MaxPolarAngle = float32(math.Pi)
	cc.MinAzimuthAngle = float32(mgl32.Inf(-1))
	cc.MaxAzimuthAngle = float32(mgl32.Inf(1))
	cc.RotSpeed = 1.0
	cc.ZoomSpeed = 0.1
	cc.KeyRotSpeed = 15 * float32(math.Pi/180) // 15 degrees as radians
	cc.KeyZoomSpeed = 2.0
	cc.KeyPanSpeed = 35.0

	return cc
}

func (cc *CameraControl) SetAreaSize(w, h float32) {
	cc.areaSize[0] = w
	cc.areaSize[1] = h
}

// Reset resets the orbit control.
func (cc *CameraControl) Reset() {
	cc.target = mgl32.Vec3{0, 0, 0}
}

// Target returns the current orbit target.
func (cc *CameraControl) Target() mgl32.Vec3 {
	return cc.target
}

//Set camera orbit target Vector3
func (cc *CameraControl) SetTarget(v mgl32.Vec3) {
	cc.target = v
}

// Enabled returns the current OrbitEnabled bitmask.
func (cc *CameraControl) Enabled() OrbitEnabled {
	return cc.enabled
}

// SetEnabled sets the current OrbitEnabled bitmask.
func (cc *CameraControl) SetEnabled(bitmask OrbitEnabled) {
	cc.enabled = bitmask
}

// Rotate rotates the camera around the target by the specified angles.
func (cc *CameraControl) Rotate(thetaDelta, phiDelta float32) {

	const EPS = 0.0001
	// Compute direction vector from target to camera
	tcam := cc.camera.Position().Sub(cc.target)

	// Calculate angles based on current camera position plus deltas
	radius := tcam.Len()
	theta := mgl32.Atan2(tcam.X(), tcam.Z()) + thetaDelta
	phi := float32(mgl32.Acos(tcam.Y()/radius)) + phiDelta

	// Restrict phi and theta to be between desired limits
	phi = mgl32.Clamp(phi, cc.MinPolarAngle, cc.MaxPolarAngle)
	phi = mgl32.Clamp(phi, EPS, float32(math.Pi)-EPS)
	theta = mgl32.Clamp(theta, cc.MinAzimuthAngle, cc.MaxAzimuthAngle)

	// Calculate new cartesian coordinates
	tcam[0] = radius * mgl32.Sin(phi) * mgl32.Sin(theta)
	tcam[1] = radius * mgl32.Cos(phi)
	tcam[2] = radius * mgl32.Sin(phi) * mgl32.Cos(theta)

	// Update camera position and orientation
	cc.camera.SetPosition(cc.target.Add(tcam))
	cc.camera.LookAt(cc.target, cc.up)
}

// Zoom moves the camera closer or farther from the target the specified amount
// and also updates the camera's orthographic size to match.
func (cc *CameraControl) Zoom(delta float32) {

	// Compute direction vector from target to camera
	tcam := cc.camera.Position().Sub(cc.target)

	// Calculate new distance from target and apply limits
	dist := tcam.Len() * (1 + delta/10)
	dist = mgl32.Max(cc.MinDistance, mgl32.Min(cc.MaxDistance, dist))
	tcam = SetLength(tcam, dist)

	// Update orthographic size and camera position with new distance
	cc.camera.UpdateSize(tcam.Len())
	cc.camera.SetPosition(cc.target.Add(tcam))
}

// Pan pans the camera and target the specified amount on the plane perpendicular to the viewing direction.
func (cc *CameraControl) Pan(deltaX, deltaY float32) {
	log.Printf("delta %f %f", deltaX, deltaY)
	// Compute direction vector from camera to target
	position := cc.camera.Position()
	vdir := cc.target.Sub(position)
	log.Printf("pos: %+v, target: %+v", position, cc.target)

	// Conversion constant between an on-screen cursor delta and its projection on the target plane
	c := 2 * vdir.Len() * mgl32.Tan(mgl32.DegToRad(cc.camera.fov/2.0)) / cc.areaSize[0]

	// Calculate pan components, scale by the converted offsets and combine them
	var pan, panX, panY mgl32.Vec3
	panX = cc.up.Cross(vdir).Normalize()
	panY = vdir.Cross(panX).Normalize()

	panX = panX.Mul(c * deltaX)
	panY = panY.Mul(c * deltaY)

	pan = panX.Add(panY)

	// Add pan offset to camera and target
	cc.camera.SetPosition(position.Add(pan))
	cc.target = cc.target.Add(pan)
}

// onMouse is called when an OnMouseDown/OnMouseUp event is received.
func (cc *CameraControl) OnMouseButton(b glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {

	// If nothing enabled ignore event
	if cc.enabled == OrbitNone {
		return
	}

	switch action {
	case glfw.Press:
		switch b {
		case glfw.MouseButtonLeft: // Rotate
			if mods&glfw.ModShift > 0 {
				if cc.enabled&OrbitPan > 0 && mods&glfw.ModShift > 0 {
					cc.state = statePan
					cc.panStart[0] = cc.mousePos[0]
					cc.panStart[1] = cc.mousePos[1]
				}
			} else {
				if cc.enabled&OrbitRot != 0 {
					cc.state = stateRotate
					cc.rotStart[0] = cc.mousePos[0]
					cc.rotStart[1] = cc.mousePos[1]
				}
			}
		case glfw.MouseButtonRight: // Pan
			if cc.enabled&OrbitPan != 0 {
				cc.state = statePan
				cc.panStart[0] = cc.mousePos[0]
				cc.panStart[1] = cc.mousePos[1]
			}
		}
	case glfw.Release:
		cc.state = stateNone
	}
}

// onCursor is called when an OnCursor event is received.
func (cc *CameraControl) OnCursorPos(xPos float64, yPos float64) {
	// If nothing enabled ignore event
	if cc.enabled == OrbitNone {
		return
	}

	switch cc.state {
	case stateNone:
		cc.mousePos[0] = float32(xPos)
		cc.mousePos[1] = float32(yPos)
	case stateRotate:
		c := -2 * math.Pi * cc.RotSpeed / cc.areaSize[0]
		cc.Rotate(
			c*(float32(xPos)-cc.rotStart.X()),
			c*(float32(yPos)-cc.rotStart.Y()),
		)
		cc.rotStart[0] = float32(xPos)
		cc.rotStart[1] = float32(yPos)
	case stateZoom:
		cc.Zoom(cc.ZoomSpeed * (float32(yPos) - cc.zoomStart))
		cc.zoomStart = float32(yPos)
	case statePan:
		cc.Pan(
			float32(xPos)-cc.panStart.X(),
			float32(yPos)-cc.panStart.Y(),
		)
		cc.panStart[0] = float32(xPos)
		cc.panStart[1] = float32(yPos)
	}
}

// onScroll is called when an OnScroll event is received.
func (cc *CameraControl) OnMouseScroll(x, y float64) {
	if cc.enabled&OrbitZoom != 0 {
		cc.Zoom(float32(-y))
	}
}

// onKey is called when an OnKeyDown/OnKeyRepeat event is received.
func (cc *CameraControl) OnKey(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// If keyboard control is disabled ignore event
	if cc.enabled&OrbitKeys == 0 {
		return
	}

	if mods == 0 && cc.enabled&OrbitRot != 0 {
		switch key {
		case glfw.KeyUp:
			cc.Rotate(0, -cc.KeyRotSpeed)
		case glfw.KeyDown:
			cc.Rotate(0, cc.KeyRotSpeed)
		case glfw.KeyLeft:
			cc.Rotate(-cc.KeyRotSpeed, 0)
		case glfw.KeyRight:
			cc.Rotate(cc.KeyRotSpeed, 0)
		}
	}
	if mods&glfw.ModControl > 0 && cc.enabled&OrbitZoom != 0 {
		switch key {
		case glfw.KeyUp:
			cc.Zoom(-cc.KeyZoomSpeed)
		case glfw.KeyDown:
			cc.Zoom(cc.KeyZoomSpeed)
		}
	}
	if mods&glfw.ModShift > 0 && cc.enabled&OrbitPan != 0 {
		switch key {
		case glfw.KeyUp:
			cc.Pan(0, cc.KeyPanSpeed)
		case glfw.KeyDown:
			cc.Pan(0, -cc.KeyPanSpeed)
		case glfw.KeyLeft:
			cc.Pan(cc.KeyPanSpeed, 0)
		case glfw.KeyRight:
			cc.Pan(-cc.KeyPanSpeed, 0)
		}
	}
}

func SetVectorFromQuat(q mgl32.Quat) mgl32.Vec3 {
	m := q.Mat4()
	return SetVectorFromRotationMatrix(m)
}

func LookAt(eye, target, up mgl32.Vec3) mgl32.Mat4 {
	z := eye.Sub(target)
	if z.LenSqr() == 0 {
		// Eye and target are in the same position
		z[2] = 1
	}
	z = z.Normalize()

	x := up.Cross(z)
	if x.LenSqr() == 0 {
		// Up and Z are parallel
		if mgl32.Abs(up.Z()) == 1 {
			z[0] += 0.0001
		} else {
			z[2] += 0.0001
		}
		z = z.Normalize()
		x = up.Cross(z)
	}
	x = x.Normalize()

	y := z.Cross(x)

	var m mgl32.Mat4
	m[0] = x.X()
	m[1] = x.Y()
	m[2] = x.Z()

	m[4] = y.X()
	m[5] = y.Y()
	m[6] = y.Z()

	m[8] = z.X()
	m[9] = z.Y()
	m[10] = z.Z()

	return m
}
func SetVectorFromRotationMatrix(r mgl32.Mat4) mgl32.Vec3 {
	m11 := r[0]
	m12 := r[4]
	m13 := r[8]
	m22 := r[5]
	m23 := r[9]
	m32 := r[6]
	m33 := r[10]

	var v mgl32.Vec3
	v[1] = float32(mgl32.Asin(mgl32.Clamp(m13, -1, 1)))
	if mgl32.Abs(m13) < 0.99999 {
		v[0] = mgl32.Atan2(-m23, m33)
		v[2] = mgl32.Atan2(-m12, m11)
	} else {
		v[0] = mgl32.Atan2(m32, m22)
		v[2] = 0
	}
	return v
}

func SetFromEuler(a mgl32.Vec3) mgl32.Quat {
	s1, c1 := mgl32.Sincos(a.X() / 2)
	s2, c2 := mgl32.Sincos(a.Y() / 2)
	s3, c3 := mgl32.Sincos(a.Z() / 2)
	ret := mgl32.Quat{}

	ret.V = mgl32.Vec3{
		float32(s1*c2*c3 - c1*s2*s3),
		float32(c1*s2*c3 + s1*c2*s3),
		float32(c1*c2*s3 - s1*s2*c3),
	}
	ret.W = float32(c1*c2*c3 + s1*s2*s3)
	return ret
}

func DecomposeMatrixToPosQuatScale(m mgl32.Mat4) (mgl32.Vec3, mgl32.Quat, mgl32.Vec3) {
	position := mgl32.Vec3{m[12], m[13], m[14]}
	scale := mgl32.Vec3{
		mgl32.Vec3{m[0], m[1], m[2]}.Len(),
		mgl32.Vec3{m[4], m[5], m[6]}.Len(),
		mgl32.Vec3{m[8], m[9], m[10]}.Len(),
	}

	// If determinant is negative, we need to invert one scale
	det := m.Det()
	if det < 0 {
		scale[0] = -scale.X()
	}

	// Scale the rotation part
	invSX := 1 / scale.X()
	invSY := 1 / scale.Y()
	invSZ := 1 / scale.Z()

	m[0] *= invSX
	m[1] *= invSX
	m[2] *= invSX

	m[4] *= invSY
	m[5] *= invSY
	m[6] *= invSY

	m[8] *= invSZ
	m[9] *= invSZ
	m[10] *= invSZ

	quaternion := mgl32.Mat4ToQuat(m)

	return position, quaternion, scale
}

func Compose(pos mgl32.Vec3, quat mgl32.Quat, scale mgl32.Vec3) mgl32.Mat4 {
	m := quat.Mat4()
	m = Scale(m, scale)
	m[12] = pos.X()
	m[13] = pos.Y()
	m[14] = pos.Z()
	return m

}

func SetPos(m mgl32.Mat4, v mgl32.Vec3) mgl32.Mat4 {
	m[12] = v.X()
	m[13] = v.Y()
	m[14] = v.Z()
	return m
}

func Scale(m mgl32.Mat4, v mgl32.Vec3) mgl32.Mat4 {
	m[0] *= v.X()
	m[4] *= v.Y()
	m[8] *= v.Z()
	m[1] *= v.X()
	m[5] *= v.Y()
	m[9] *= v.Z()
	m[2] *= v.X()
	m[6] *= v.Y()
	m[10] *= v.Z()
	m[3] *= v.X()
	m[7] *= v.Y()
	m[11] *= v.Z()
	return m
}

func SetLength(v mgl32.Vec3, l float32) mgl32.Vec3 {
	oldLength := v.Len()
	if oldLength != 0 && l != oldLength {
		return v.Mul(l / oldLength)
	}
	return v
}
