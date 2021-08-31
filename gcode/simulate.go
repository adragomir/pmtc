package gcode

import (
	"errors"
	"math"
	"reflect"

	"github.com/adragomir/linuxcncgo/gcode/parser"
	"github.com/adragomir/linuxcncgo/ui/glutil"
	"github.com/go-gl/mathgl/mgl32"
)

type Plane struct {
	firstCoord        int
	secondCoord       int
	lastCoord         int
	firstCenterCoord  string
	secondCenterCoord string
}

var (
	XY_PLANE = Plane{
		firstCoord:        0,
		secondCoord:       1,
		lastCoord:         2,
		firstCenterCoord:  "I",
		secondCenterCoord: "J",
	}

	YZ_PLANE = Plane{
		firstCoord:        1,
		secondCoord:       2,
		lastCoord:         0,
		firstCenterCoord:  "J",
		secondCenterCoord: "K",
	}

	XZ_PLANE = Plane{
		firstCoord:        2,
		secondCoord:       0,
		lastCoord:         1,
		firstCenterCoord:  "K",
		secondCenterCoord: "I",
	}
)

type FragmentType int

const (
	LineFragmentType FragmentType = iota
	ArcFragmentType
)

type SpeedTagType int

const (
	NormalSpeedTag SpeedTagType = iota
	RapidSpeedTag
)

type SpeedType int

const (
	SpeedAccel SpeedType = iota
	SpeedDecel
	SpeedConst
)

type SpeedData struct {
	length float32
}

type RunFragment struct {
	tp           SpeedType
	fragment     *Fragment
	length       float32
	duration     float32
	fromSqSpeed  float32
	toSqSpeed    float32
	squaredSpeed float32
	startX       float32
	stopX        float32
}

type Fragment struct {
	tp              FragmentType
	from            mgl32.Vec3
	to              mgl32.Vec3
	plane           Plane
	center          mgl32.Vec2
	centerInPlane   mgl32.Vec2
	fromAngle       float32
	angularDistance float32
	radius          float32
	feedRate        float32
	lineNo          int
	speedTag        SpeedTagType

	length       float32
	duration     float32
	squaredSpeed float32
	maxAccel     float32

	RunData      map[SpeedType]SpeedData
	RunFragments []*RunFragment
}

type MachineState struct {
	position       mgl32.Vec3
	lineNo         int
	distanceMode   func(mgl32.Vec3, Move) mgl32.Vec3
	motionMode     func([]parser.Expr, *MachineState)
	planeMode      Plane
	feedRate       float32
	travelFeedRate float32
	pathControl    string
	path           []*Fragment
	origins        []mgl32.Vec3
	currentOrigin  int

	accumulator *Accumulator
	fragmentCb  func(*Fragment)
}

var GROUPS_TRANSITIONS = map[string]interface{}{
	"0":    map[string]interface{}{"motionMode": moveG0},
	"1":    map[string]interface{}{"motionMode": moveG1},
	"2":    map[string]interface{}{"motionMode": moveG2},
	"3":    map[string]interface{}{"motionMode": moveG3},
	"4":    nil,
	"17":   map[string]interface{}{"planeMode": XY_PLANE},
	"18":   map[string]interface{}{"planeMode": XZ_PLANE},
	"19":   map[string]interface{}{"planeMode": YZ_PLANE},
	"20":   nil,
	"21":   "unsupported",
	"40":   "unsupported",
	"41":   "unsupported",
	"42":   "unsupported",
	"49":   nil,
	"54":   map[string]interface{}{"currentOrigin": 1},
	"55":   map[string]interface{}{"currentOrigin": 2},
	"56":   map[string]interface{}{"currentOrigin": 3},
	"57":   map[string]interface{}{"currentOrigin": 4},
	"58":   map[string]interface{}{"currentOrigin": 5},
	"59":   map[string]interface{}{"currentOrigin": 6},
	"59.1": map[string]interface{}{"currentOrigin": 7},
	"59.2": map[string]interface{}{"currentOrigin": 8},
	"59.3": map[string]interface{}{"currentOrigin": 9},
	"61":   map[string]interface{}{"pathControl": "61"},
	"61.1": map[string]interface{}{"pathControl": "61.1"},
	"64":   map[string]interface{}{"pathControl": "64"},
	"80":   map[string]interface{}{"motionMode": moveG80},
	"90":   map[string]interface{}{"distanceMode": absoluteDistanceMode},
	"91":   map[string]interface{}{"distanceMode": incrementalDistanceMode},
	"94":   nil,
}

func NewMachineState(feedRate float32, travelFeedRate float32, initialPosition mgl32.Vec3) *MachineState {
	origins := make([]mgl32.Vec3, 10)
	for i := 0; i < 10; i++ {
		origins[i] = mgl32.Vec3{0.0, 0.0, 0.0}
	}
	accumulator := NewAccumulator()
	tmp := &MachineState{
		position:       initialPosition,
		distanceMode:   absoluteDistanceMode,
		motionMode:     moveG0,
		planeMode:      XY_PLANE,
		feedRate:       feedRate,
		travelFeedRate: travelFeedRate,
		pathControl:    "61",
		path:           make([]*Fragment, 0),
		origins:        origins,
		currentOrigin:  1,

		accumulator: accumulator,
		fragmentCb:  accumulator.FragmentListener,
	}
	return tmp
}

func (ms *MachineState) absolutePoint(in Move) mgl32.Vec3 {
	currentOrigin := ms.origins[ms.currentOrigin]
	out := Move{}
	for i, key := range []string{"X", "Y", "Z"} {
		if val, ok := in[key]; ok {
			out[key] = val + currentOrigin[i]
		}
	}
	return ms.distanceMode(ms.position, out)
}

func (ms *MachineState) addPathFragment(p *Fragment) {
	ms.path = append(ms.path, p)
	if ms.fragmentCb != nil {
		ms.fragmentCb(p)
	}
}

type Move map[string]float32

func moveG0(line []parser.Expr, ms *MachineState) {
	moveStraight(line, ms, ms.travelFeedRate, RapidSpeedTag)
}

func moveG1(line []parser.Expr, ms *MachineState) {
	moveStraight(line, ms, ms.feedRate, NormalSpeedTag)
}

func moveG2(line []parser.Expr, ms *MachineState) {
	moveArc(line, true, ms)
}

func moveG3(line []parser.Expr, ms *MachineState) {
	moveArc(line, false, ms)
}

func mapify(line []parser.Expr) (map[string]float32, map[string]string) {
	out := make(map[string]float32)
	outString := make(map[string]string)
	for _, tmp := range line {
		switch tmp.(type) {
		case parser.WordExpr:
			we := tmp.(parser.WordExpr)
			out[we.Word] = float32(we.Val.(parser.ConstExpr).Val)
			outString[we.Word] = we.Val.(parser.ConstExpr).RawVal
		}
	}
	return out, outString
}
func detectMove(line []parser.Expr) Move {
	result := Move{}
	for _, tmp := range line {
		switch tmp.(type) {
		case parser.WordExpr:
			we := tmp.(parser.WordExpr)
			switch we.Word {
			case "X":
				result["X"] = float32(we.Val.(parser.ConstExpr).Val)
			case "Y":
				result["Y"] = float32(we.Val.(parser.ConstExpr).Val)
			case "Z":
				result["Z"] = float32(we.Val.(parser.ConstExpr).Val)
			}
		}
	}
	return result
}

func moveArc(line []parser.Expr, clockwise bool, ms *MachineState) {
	move := detectMove(line)
	if len(move) == 0 {
		return
	}

	currentPosition := ms.position
	targetPos := ms.absolutePoint(move)
	plane := ms.planeMode
	xCoord := plane.firstCoord
	yCoord := plane.secondCoord
	radius, toCenterX, toCenterY, err := findCircle(line, targetPos, plane, currentPosition, clockwise)
	if err == nil {
		centerX := currentPosition[xCoord] + toCenterX
		centerY := currentPosition[yCoord] + toCenterY
		targetCenterX := targetPos[xCoord] - centerX
		targetCenterY := targetPos[yCoord] - centerY
		angularDiff := mgl32.Atan2(
			-toCenterX*targetCenterY+toCenterY*targetCenterX,
			-toCenterX*targetCenterX-toCenterY*targetCenterY,
		)
		if clockwise && angularDiff >= 0 {
			angularDiff -= 2 * math.Pi
		}
		if !clockwise && angularDiff <= 0 {
			angularDiff += 2 * math.Pi
		}
		angularStart := mgl32.Atan2(-toCenterY, -toCenterX)
		ms.addPathFragment(&Fragment{
			tp:              ArcFragmentType,
			from:            currentPosition,
			to:              targetPos,
			plane:           plane,
			center:          mgl32.Vec2{centerX, centerY},
			centerInPlane:   mgl32.Vec2{centerX, centerY},
			fromAngle:       angularStart,
			angularDistance: angularDiff,
			radius:          radius,
			feedRate:        ms.feedRate,
			lineNo:          ms.lineNo,
			speedTag:        NormalSpeedTag,

			RunFragments: make([]*RunFragment, 0),
			RunData:      make(map[SpeedType]SpeedData),
		})
		ms.position = targetPos
	}
}

func findCircle(line []parser.Expr, targetPos mgl32.Vec3, plane Plane, currentPosition mgl32.Vec3, clockwise bool) (float32, float32, float32, error) {
	lineMap, _ := mapify(line)
	var radius, toCenterX, toCenterY float32
	var err error
	if tmp, ok := lineMap["R"]; ok {
		radius = tmp
		dx := targetPos[plane.firstCoord] - currentPosition[plane.firstCoord]
		dy := targetPos[plane.secondCoord] - currentPosition[plane.secondCoord]
		mightyFactor := 4*radius*radius - dx*dx - dy*dy
		mightyFactor = -mgl32.Sqrt(mightyFactor) / mgl32.Sqrt(dx*dx+dy*dy)
		if !clockwise {
			mightyFactor = -mightyFactor
		}
		if radius < 0 {
			mightyFactor = -mightyFactor
			radius = -radius
		}
		toCenterX = 0.5 * (dx - (dy * mightyFactor))
		toCenterY = 0.5 * (dy + (dx * mightyFactor))
	} else {
		xMatch, okx := lineMap[plane.firstCenterCoord]
		yMatch, oky := lineMap[plane.secondCenterCoord]
		if !okx && !oky {
			err = errors.New("no center")
		}
		if okx {
			toCenterX = xMatch
		} else {
			toCenterX = 0
		}
		if oky {
			toCenterY = yMatch
		} else {
			toCenterY = 0
		}
		radius = mgl32.Sqrt(toCenterX*toCenterX + toCenterY*toCenterY)
	}
	return radius, toCenterX, toCenterY, err
}

func moveStraight(line []parser.Expr, ms *MachineState, feedRate float32, speedTag SpeedTagType) {
	move := detectMove(line)
	if len(move) > 0 {
		apoint := ms.absolutePoint(move)
		addPathComponent(apoint, ms, feedRate, speedTag)
	}
}

func addPathComponent(apoint mgl32.Vec3, ms *MachineState, feedRate float32, speedTag SpeedTagType) {
	hasMove := false
	for i := 0; i < 3; i++ {
		hasMove = hasMove || mgl32.Abs(apoint[i]-ms.position[i]) > 0.00001
	}
	if hasMove {
		ms.addPathFragment(&Fragment{
			tp:       LineFragmentType,
			from:     ms.position,
			to:       apoint,
			feedRate: feedRate,
			lineNo:   ms.lineNo,
			speedTag: speedTag,

			RunFragments: make([]*RunFragment, 0),
			RunData:      make(map[SpeedType]SpeedData),
		})
		ms.position = apoint
	}
}

func moveG80(line []parser.Expr, ms *MachineState) {
	// do Nothing
}

func absoluteDistanceMode(pos mgl32.Vec3, move Move) mgl32.Vec3 {
	if val, ok := move["X"]; ok {
		pos[0] = val
	}
	if val, ok := move["Y"]; ok {
		pos[1] = val
	}
	if val, ok := move["Z"]; ok {
		pos[2] = val
	}
	return pos
}

func incrementalDistanceMode(pos mgl32.Vec3, move Move) mgl32.Vec3 {
	if val, ok := move["X"]; ok {
		pos[0] = pos[0] + val
	}
	if val, ok := move["Y"]; ok {
		pos[1] = pos[1] + val
	}
	if val, ok := move["Z"]; ok {
		pos[2] = pos[2] + val
	}
	return pos
}

type SimFragment struct {
	Vertices []mgl32.Vec3
	SpeedTag SpeedTagType
}

type SimMap struct {
	Segments []mgl32.Vec3
	LineNo   int
}

type Accumulator struct {
	Simulation   []*SimFragment
	currSpeedTag SpeedTagType
	CurrPath     []mgl32.Vec3
	SegmentMap   []SimMap
	handler      func(*SimFragment)
}

func NewAccumulator() *Accumulator {
	return &Accumulator{
		Simulation: make([]*SimFragment, 0),
		CurrPath:   make([]mgl32.Vec3, 0),
	}
}

func (a *Accumulator) Close() {
	if len(a.CurrPath) > 0 {
		tmp := &SimFragment{
			Vertices: make([]mgl32.Vec3, len(a.CurrPath)),
			SpeedTag: a.currSpeedTag,
		}
		for i := range a.CurrPath {
			tmp.Vertices[i] = a.CurrPath[i]
		}
		a.Simulation = append(a.Simulation, tmp)
		lastPoint := a.CurrPath[len(a.CurrPath)-1]
		a.CurrPath = []mgl32.Vec3{lastPoint}
		if a.handler != nil {
			a.handler(tmp)
		}
	}
}

func (a *Accumulator) Accumulate(p mgl32.Vec3, speedTag SpeedTagType) {
	if a.currSpeedTag != speedTag || len(a.CurrPath) >= 10000 {
		a.Close()
		a.currSpeedTag = speedTag
	}
	a.CurrPath = append(a.CurrPath, p)
}

func (a *Accumulator) Empty() bool {
	return len(a.Simulation) == 0 && len(a.CurrPath) == 0
}

func (a *Accumulator) FragmentListener(f *Fragment) {
	if a.Empty() {
		a.Accumulate(f.from, f.speedTag)
	}
	if f.tp == LineFragmentType {
		tmp := SimMap{
			Segments: []mgl32.Vec3{f.from, f.to},
			LineNo:   f.lineNo,
		}
		a.SegmentMap = append(a.SegmentMap, tmp)
		a.Accumulate(f.to, f.speedTag)
	} else {
		tolerance := float32(0.0001)
		steps := mgl32.Ceil(
			math.Pi /
				mgl32.Acos(1-tolerance/f.radius) * mgl32.Abs(f.angularDistance) /
				(math.Pi * 2),
		)
		tmp := SimMap{
			Segments: []mgl32.Vec3{},
			LineNo:   f.lineNo,
		}
		points := []mgl32.Vec3{}
		for j := 0; j <= int(steps); j++ {
			point := COMPONENT_TYPES[ArcFragmentType].PointAtRatio(f, float32(j)/steps)
			points = append(points, point)
			a.Accumulate(point, f.speedTag)
		}
		tmp.Segments = append(tmp.Segments, points...)
		a.SegmentMap = append(a.SegmentMap, tmp)
	}
}

func Evaluate(in string, feedRate float32, travelFeedRate float32, initialPosition mgl32.Vec3) ([]*Fragment, *Accumulator) {
	ms := NewMachineState(feedRate, travelFeedRate, initialPosition)
	parsed, _ := parser.ParseAll(in)
	for _, line := range parsed {
		lineNo := line.LineNo
		ms.lineNo = lineNo
		ast := line.Ast
		astMap, astMapString := mapify(ast)
		if f, ok := astMap["F"]; ok {
			ms.feedRate = f
		}
		if g, ok := astMapString["G"]; ok {
			if rawTrans, ok := GROUPS_TRANSITIONS[g]; ok && rawTrans != nil {
				switch trans := rawTrans.(type) {
				case string:
					// unsupported
				case map[string]interface{}:
					for k, v := range trans {
						switch k {
						case "motionMode":
							ms.motionMode = v.(func([]parser.Expr, *MachineState))
						case "currentOrigin":
							ms.currentOrigin = v.(int)
						case "planeMode":
							ms.planeMode = v.(Plane)
						case "pathControl":
							ms.pathControl = v.(string)
						case "distanceMode":
							ms.distanceMode = v.(func(mgl32.Vec3, Move) mgl32.Vec3)

						}
					}
				}
			}
		}
		ms.motionMode(line.Ast, ms)
	}
	return ms.path, ms.accumulator
}

func Simulate(in string) {
}

type ComponentType interface {
	Length(f *Fragment) float32
	Speed(f *Fragment, accel float32) (float32, float32)
	EntryDir(f *Fragment) mgl32.Vec3
	ExitDir(f *Fragment) mgl32.Vec3
	PointAtRatio(f *Fragment, ratio float32) mgl32.Vec3
	SimSteps(f *Fragment) int
}

type LineComponent struct {
}

func (l *LineComponent) Length(f *Fragment) float32 {
	return f.to.Sub(f.from).Len()
}
func (l *LineComponent) Speed(f *Fragment, accel float32) (float32, float32) {
	return f.feedRate / 60.0, accel
}
func (l *LineComponent) EntryDir(f *Fragment) mgl32.Vec3 {
	return f.to.Sub(f.from).Normalize()
}
func (l *LineComponent) ExitDir(f *Fragment) mgl32.Vec3 {
	return f.to.Sub(f.from).Normalize()
}
func (l *LineComponent) PointAtRatio(f *Fragment, ratio float32) mgl32.Vec3 {
	return f.from.Add(f.to.Sub(f.from).Mul(ratio))
}
func (l *LineComponent) SimSteps(f *Fragment) int {
	return 40
}

type ArcComponent struct {
}

func (l *ArcComponent) Length(f *Fragment) float32 {
	lastCoord := f.plane.lastCoord
	planarArcLength := f.angularDistance * f.radius
	return mgl32.Sqrt(
		mgl32.Pow(planarArcLength, 2) +
			mgl32.Pow(f.to[lastCoord]-f.from[lastCoord], 2),
	)
}

func (l *ArcComponent) Speed(f *Fragment, accel float32) (float32, float32) {
	radius := f.radius
	speed := f.feedRate / 60.0
	maxRadialAccel := mgl32.Pow(speed, 2) / radius
	reductionFactor := float32(0.8)
	if maxRadialAccel > accel*reductionFactor {
		// constant speed would already create a too big radial acceleration, reducing speed.
		speed = mgl32.Sqrt(accel * reductionFactor * radius)
		maxRadialAccel = mgl32.Pow(speed, 2) / radius
	}
	maxTangentialAccel := mgl32.Sqrt(mgl32.Pow(accel, 2) - mgl32.Pow(maxRadialAccel, 2))
	return speed, maxTangentialAccel
}

func getArcSpeedDir(f *Fragment, angle float32) mgl32.Vec3 {
	p := f.plane
	angle = 0
	rx := mgl32.Cos(f.fromAngle + angle)
	ry := mgl32.Sin(f.fromAngle + angle)
	dz := (f.to[p.lastCoord] - f.from[p.lastCoord]) / mgl32.Abs(f.angularDistance) / f.radius
	lgt := mgl32.Vec3{rx, ry, dz}.Len()
	var dir float32
	if f.angularDistance >= 0 {
		dir = 1.0
	} else {
		dir = -1.0
	}
	return mgl32.Vec3{-dir * ry / lgt, dir * rx / lgt, dz / lgt}
}

func (l *ArcComponent) EntryDir(f *Fragment) mgl32.Vec3 {
	return getArcSpeedDir(f, 0)
}

func (l *ArcComponent) ExitDir(f *Fragment) mgl32.Vec3 {
	return getArcSpeedDir(f, f.angularDistance)
}

func polarPoint(r float32, theta float32) mgl32.Vec2 {
	return mgl32.Vec2{
		r * mgl32.Cos(theta), r * mgl32.Sin(theta),
	}
}

func (l *ArcComponent) PointAtRatio(f *Fragment, ratio float32) mgl32.Vec3 {
	pointInPlane := f.centerInPlane.Add(polarPoint(f.radius, f.fromAngle+f.angularDistance*ratio))
	np := mgl32.Vec3{}
	np[f.plane.firstCoord] = pointInPlane[0]
	np[f.plane.secondCoord] = pointInPlane[1]
	np[f.plane.lastCoord] = (f.from[f.plane.lastCoord]*(1-ratio) + f.to[f.plane.lastCoord]*ratio)
	return np
}

func (l *ArcComponent) SimSteps(f *Fragment) int {
	return int(mgl32.Round(mgl32.Abs(f.angularDistance)/(2*math.Pi)*50, 3))
}

var COMPONENT_TYPES = map[FragmentType]ComponentType{
	LineFragmentType: &LineComponent{},
	ArcFragmentType:  &ArcComponent{},
}

func groupConnectedComponents(path []*Fragment, accel float32) [][]*Fragment {
	groups := make([][]*Fragment, 0)
	lastExitDir := mgl32.Vec3{0, 0, 0}
	var currentGroup []*Fragment
	for _, component := range path {
		trait := COMPONENT_TYPES[component.tp]
		// are mostly continuous
		cutOff := float32(1.95)
		if lastExitDir.Add(trait.EntryDir(component)).LenSqr() >= cutOff*cutOff {
			currentGroup = append(currentGroup, component)
		} else {
			currentGroup := []*Fragment{component}
			groups = append(groups, currentGroup)
		}
		lastExitDir = trait.ExitDir(component)
		sp, acc := trait.Speed(component, accel)
		component.length = trait.Length(component)
		component.squaredSpeed, component.maxAccel = mgl32.Pow(sp, 2), acc
	}
	return groups
}

type SimulateStats struct {
	TotalTime float32
	Bbox      *glutil.BoundingBox
}

func (s *SimulateStats) Push(x, y, z, t float32) {
	s.TotalTime = mgl32.Max(s.TotalTime, t)
	s.Bbox.PushCoord(x, y, z)
}

type Simulation struct {
	lastPos     mgl32.Vec3
	CurrentTime float32
	stats       SimulateStats
}

var FRAGMENT_EQUATIONS map[SpeedType](func(*RunFragment, float32, float32) (float32, float32)) = map[SpeedType]func(*RunFragment, float32, float32) (float32, float32){
	SpeedAccel: func(rf *RunFragment, ratio float32, accel float32) (float32, float32) {
		x2 := 2*rf.fragment.RunFragments[SpeedAccel].length + rf.length*ratio
		return mgl32.Sqrt(accel * x2), mgl32.Sqrt(x2/accel) - mgl32.Sqrt(rf.fromSqSpeed)/accel
	},
	SpeedDecel: func(rf *RunFragment, ratio float32, accel float32) (float32, float32) {
		x2 := 2 * (rf.fragment.RunFragments[SpeedDecel].length + rf.length*(1-ratio))
		return mgl32.Sqrt(accel * x2), mgl32.Sqrt(rf.toSqSpeed)/accel + rf.duration - mgl32.Sqrt(x2/accel)
	},
	SpeedConst: func(rf *RunFragment, ratio float32, accel float32) (float32, float32) {
		return mgl32.Sqrt(rf.squaredSpeed), rf.duration * ratio
	},
}

func dataForRatio(f *Fragment, ratio float32) (float32, float32) {
	acceleration := f.maxAccel
	x := f.length * ratio
	timeOffset := float32(0.0)
	xOffset := float32(0.0)
	fragmentIndex := 0
	fragment := f.RunFragments[fragmentIndex]
	for fragment.stopX < x {
		timeOffset += fragment.duration
		xOffset += fragment.length
		fragmentIndex++
		fragment = f.RunFragments[fragmentIndex]
	}
	speed, time := FRAGMENT_EQUATIONS[fragment.tp](fragment, mgl32.Min(1, (x-xOffset)/fragment.length), acceleration)
	return speed, timeOffset + time
}

func (s *Simulation) Discretize(f *Fragment) {
	trait := COMPONENT_TYPES[f.tp]
	steps := trait.SimSteps(f)
	startTime := s.CurrentTime
	for j := 1; j <= steps; j++ {
		ratio := float32(j / steps)
		_, time := dataForRatio(f, ratio)
		s.CurrentTime = startTime + time
		s.Push(trait.PointAtRatio(f, ratio), f)
	}
}

func (s *Simulation) Push(p mgl32.Vec3, f *Fragment) {
	s.lastPos = p
	s.stats.Push(p[0], p[1], p[2], s.CurrentTime)
}

func reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func limitSpeed(fragments []*Fragment, direction SpeedType) {
	for i := 0; i < len(fragments); i++ {
		fragment := fragments[i]
		acceleration := fragment.maxAccel
		previousSquaredSpeed := float32(0.0)
		if i > 0 {
			previousSquaredSpeed = fragments[i-1].squaredSpeed
		}
		accelerationLength := previousSquaredSpeed / (2 * acceleration)
		maxSquaredSpeed := 2 * acceleration * (accelerationLength + fragment.length)
		fragment.squaredSpeed = mgl32.Min(fragment.squaredSpeed, maxSquaredSpeed)
		fragment.RunData[direction] = SpeedData{
			length: previousSquaredSpeed / (2 * acceleration),
		}
	}
}

func planSpeed(group []*Fragment) {
	limitSpeed(group, SpeedAccel)
	reverse(group)
	limitSpeed(group, SpeedDecel)
	reverse(group)
	for i := 0; i < len(group); i++ {
		var nextSquaredSpeed float32
		if i < len(group)-1 {
			nextSquaredSpeed = group[i+1].squaredSpeed
		} else {
			nextSquaredSpeed = 0
		}
		var previousSquaredSpeed float32
		if i >= 1 {
			previousSquaredSpeed = group[i-1].squaredSpeed
		} else {
			previousSquaredSpeed = 0
		}
		fragment := group[i]
		fragment.RunFragments = make([]*RunFragment, 0)
		acceleration := fragment.maxAccel
		accelerationLength := fragment.RunData[SpeedAccel].length
		decelerationLength := fragment.RunData[SpeedDecel].length
		meetingPoint := (decelerationLength + fragment.length - accelerationLength) / 2
		meetingSquaredSpeed := 2 * acceleration * (accelerationLength + meetingPoint)
		endAccelerationPoint := (fragment.squaredSpeed - 2*acceleration*accelerationLength) / (2 * acceleration)
		startDecelerationPoint := (2*acceleration*(decelerationLength+fragment.length) - fragment.squaredSpeed) / (2 * acceleration)
		maxSquaredSpeed := fragment.squaredSpeed
		if meetingPoint >= 0 && meetingPoint <= fragment.length && meetingSquaredSpeed <= fragment.squaredSpeed {
			maxSquaredSpeed = meetingSquaredSpeed
			endAccelerationPoint = meetingPoint
			startDecelerationPoint = meetingPoint
			fragment.squaredSpeed = meetingSquaredSpeed
		}
		hasAcceleration := endAccelerationPoint > 0 && endAccelerationPoint <= fragment.length
		hasDeceleration := startDecelerationPoint >= 0 && startDecelerationPoint < fragment.length
		if hasAcceleration {
			runFragment := &RunFragment{
				tp:          SpeedAccel,
				fragment:    fragment,
				fromSqSpeed: previousSquaredSpeed,
				toSqSpeed:   maxSquaredSpeed,
				startX:      0,
				stopX:       endAccelerationPoint,
			}
			fragment.RunFragments = append(fragment.RunFragments, runFragment)
		}
		var constantSpeedStart float32
		if hasAcceleration {
			constantSpeedStart = endAccelerationPoint
		} else {
			constantSpeedStart = 0
		}
		var constantSpeedStop float32
		if hasDeceleration {
			constantSpeedStop = startDecelerationPoint
		} else {
			constantSpeedStop = fragment.length
		}
		if constantSpeedStart != constantSpeedStop {
			fragment.RunFragments = append(fragment.RunFragments, &RunFragment{
				tp:           SpeedConst,
				fragment:     fragment,
				squaredSpeed: maxSquaredSpeed,
				startX:       constantSpeedStart,
				stopX:        constantSpeedStop,
			})
		}
		if hasDeceleration {
			fragment.RunFragments = append(fragment.RunFragments, &RunFragment{
				tp:          SpeedDecel,
				fragment:    fragment,
				fromSqSpeed: maxSquaredSpeed,
				toSqSpeed:   nextSquaredSpeed,
				startX:      startDecelerationPoint,
				stopX:       fragment.length,
			})
		}
		fragment.duration = 0.0
		for _, rf := range fragment.RunFragments {
			rf.length = rf.stopX - rf.startX
			if rf.tp == SpeedConst {
				rf.duration = fragment.length / mgl32.Sqrt(fragment.squaredSpeed)
			} else {
				rf.duration = mgl32.Abs(
					mgl32.Sqrt(rf.fromSqSpeed)-mgl32.Sqrt(rf.toSqSpeed),
				) / acceleration
			}
			fragment.duration += rf.duration
		}
	}
}

func (s *Simulation) Simulate(in string, feedRate float32, travelFeedRate float32, pos mgl32.Vec3) ([][]*Fragment, *Accumulator, SimulateStats) {
	s.stats = SimulateStats{
		TotalTime: 0.0,
		Bbox:      glutil.NewBoundingBox(),
	}

	toolPath, accumulator := Evaluate(in, feedRate, travelFeedRate, pos)
	accumulator.Close()

	if len(toolPath) > 0 {
		accel := float32(1000.0) // mm/s^2
		groups := groupConnectedComponents(toolPath, accel)
		s.CurrentTime = 0.0
		s.lastPos = mgl32.Vec3{0, 0, 0}
		for _, group := range groups {
			planSpeed(group)
			for _, fragment := range group {
				s.Discretize(fragment)
			}
			for i := 0; i < 10; i++ {
				s.CurrentTime += 0.001
				s.Push(s.lastPos, group[len(group)-1])
			}
		}
		return groups, accumulator, s.stats
	}
	return nil, nil, SimulateStats{}

}

func SimulateGCode(in string) ([][]*Fragment, *Accumulator, SimulateStats) {
	sim := Simulation{
		CurrentTime: 0.0,
	}
	groups, accumulator, stats := sim.Simulate(in, 15*60, 15*60, mgl32.Vec3{0, 0, 0})
	return groups, accumulator, stats
}

func BuildVertexData(fragments [][]*Fragment, accumulator *Accumulator, stats SimulateStats, prepare bool) ([]float32, *glutil.BoundingBox) {
	tmp := make([]float32, 0)

	for _, sf := range accumulator.Simulation {
		for _, vertex := range sf.Vertices {
			if prepare {
				tmp = append(tmp, vertex[0], vertex[2], -vertex[1])
			} else {
				tmp = append(tmp, vertex[0], vertex[1], vertex[2])
			}
			tmp = append(tmp, float32(sf.SpeedTag))
		}

	}
	return tmp, stats.Bbox
}
