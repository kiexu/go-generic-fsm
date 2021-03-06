# Go-Generic-FSM

_**A Generic Finite State Machine Implemented In Go Language**_

## Install

Go 1.18 or higher is required for generic.

```
go get github.com/kiexu/go-generic-fsm 
```

## Install Power Pack (Optional)

### Visualization Pack

[Visualization Pack](https://github.com/kiexu/go-generic-fsm-visual-pack) can launch an HTTP service on a user-specified port and return a visualized FSM graph.

```
go get github.com/kiexu/go-generic-fsm-visual-pack
```

```go
_ = fsmv.InitFSMVisualPack(&fsmv.Config{
    Port:         9527, // Customizable port
    NativeScript: true, // Users who live in poor network may need to set true
})

// Start one FSM's visualization with Visualize()
w := &fsm.VisualOpenWrapper{} 
err = demoFsm.OpenVisualization(w) // After calling Visualize(), you can get full HTTP path from w
```

After calling OpenVisualization(), make a GET request to URL: `localhost:9527(port in config)` + w.Path, 
to see the visualized state machine, current state, node index, etc:

e.g. http://localhost:9527/fsm/visualize/8f101b6e-ed2d-420a-b0d8-e7684e130a5a

```mermaid
graph RL
    0["state: [paid]<br>idx: [0]<br>[current]"]:::currentBlock -- "[deliverEvent]" --> 2["state: [done]<br>idx: [2]"]:::block
    0["state: [paid]<br>idx: [0]<br>[current]"]:::currentBlock -- "[cancelEvent]" --> 3["state: [canceled]<br>idx: [3]"]:::block
    1["state: [initial]<br>idx: [1]<br>[previous]"]:::block -- "[payEvent]" --> 0["state: [paid]<br>idx: [0]<br>[current]"]:::currentBlock
    2["state: [done]<br>idx: [2]"]:::block -- "[readyEvent]" --> 1["state: [initial]<br>idx: [1]<br>[previous]"]:::block
    3["state: [canceled]<br>idx: [3]"]:::block -- "[readyEvent]" --> 1["state: [initial]<br>idx: [1]<br>[previous]"]:::block
classDef block fill:#fdf9ee,stroke:#939391,stroke-width:2px
classDef currentBlock fill:#eee5f8,stroke:#939391,stroke-width:3px

```

**DO NOT** forget to call CloseVisualization() to release resources if you want FSM to be GC.

```go
err = demoFsm.CloseVisualization(&fsm.VisualCloseWrapper{Token: w.Token})
```

## Usage

1. Define FSM with config: `fsm.DefConfig` based on state migration;
2. Use `fsm.NewFsm(config, initStatus)` to New a `fsm.FSM`;
3. Call `fsm.FSM.Trigger()` to trigger event and run callback functions automatically.

## Demo
Let's use an FSM that simulates the online shopping process as a demo.

We want to send an SMS to the user once the status is changed. The SMS content is shown in the `Graph`'s `Edges`, and it is considered as an attribute of an `Event`.

```mermaid
flowchart LR
    initial-->|payEvent\nSMS:Thanks|paid
    paid -->|deliverEvent\nSMS:Coming| done
    paid -->|cancelEvent\nSMS:CancelOK| canceled
    canceled -->|readyEvent\nSMS:ResetOK| initial
    done -->|readyEvent\nSMS:ResetOK| initial
     
```

### Define Configuration

#### Type Definition

Start with `fsm.DefConfig`, which is the only config struct now:

```go
// DefConfig Default factory with basic config struct
// As a regular FSM, {stateVal, eventVal} need to be unique
type DefConfig[T, S comparable, U, V any] struct {
    DescList     []*DescCell[T, S, U, V] // Required. Describe FSM graph
    StatusValMap map[T]V                 // Optional. Store custom value in abstract status
}
```

You can also develop your own config struct by implementing the `fsm.Config` interface.

First, we need to decide the concrete type of the generic `[T, S, U, V]`:

|       | Desc                   | Generic type     | Required or not | Type in this demo          |
|-------|------------------------|------------------|-----------------|----------------------------|
| **T** | State type             | comparable       | `required`      | `string`(E.g "initial")    |
| **S** | Event type             | comparable       | `required`      | `string`(E.g "payEvent")   |
| **U** | Object stored in Event | any(interface{}) | `optional`      | `string`(SMS content)      |
| **V** | Object stored in State | any(interface{}) | `optional`      | `fsm.NA`(type placeholder) |

You can use `fsm.NA` to temporarily fill unused type slot as per `context.TODO`.

```go
// NA placeholder of unused type
type NA struct{}
```

#### Define DefConfig

In this demo we ignore `DefConfig.StatusValMap` because the custom attributes of the `State` are not used (filled with `fsm.NA`).

Please note if `U` or `V` is set, both of them can be easily accessed in the **callback function** or **event trigger function result**.

In below is the final config:

```go
var demoFac = &fsm.DefConfig[string, string, string, fsm.NA]{
	DescList: []*fsm.DescCell[string, string, string, fsm.NA]{
		{
			EventVal:      "payEvent",
			FromState:     []string{"initial"}, 
			ToState:       "paid",              
			EventStoreVal: "Thanks",            // SMS message
		},
		{
			EventVal:      "deliverEvent",
			FromState:     []string{"paid"},
			ToState:       "done",
			EventStoreVal: "Coming",
		},
		{
			EventVal:      "readyEvent",
			FromState:     []string{"done", "canceled"}, // Multiple fromState leads to one toState supported
			ToState:       "initial",
			EventStoreVal: "ResetOK",
		},
		{
			EventVal:      "cancelEvent",
			FromState:     []string{"paid"},
			ToState:       "canceled",
			EventStoreVal: "CancelOK",
		},
	},
}
```

#### Initialize FSM

We initialize with the above config `demoFac` and initial state `"initial"`:

```go
demoFsm, err := fsm.NewFsm[string, string, string, fsm.NA](demoFac, "initial")
```

We get the generic `FSM` **demoFsm** successfully now.

#### Trigger FSM

Use `Trigger()` to trigger event, then check the returning `Event` to check results.

```go
event, err := demoFsm.Trigger("payEvent")

// Event packaging an eventE
type Event[T, S comparable, U, V any] struct {
    fSM      *FSM[T, S, U, V]  // Pointer to fSM
    eventVal S                 // raw input event value
    args     []interface{}     // Args to pass to Callbacks
    eventE   *Edge[T, S, U, V] // An Edge for advanced access
}

// FromState get old State of FSM
func (e *Event[T, S, U, V]) FromState() (resp T)

// ToState get new State of FSM
func (e *Event[T, S, U, V]) ToState() (resp T)
```

In concurrent environment, it is strongly recommended to use the `State` in the returning `Event`, instead of the `State` of the `FSM`, because the state in the Event is immutable.

#### Other Methods

There are some other practical methods:

```go
// CanTrigger Whether given eventVal can trigger event
func (f *FSM[T, S, U, V]) CanTrigger(eventVal S) bool

// PeekState Peek a state by prev state and event
func (f *FSM[T, S, U, V]) PeekState(state T, eventVal S) (T, bool)

// CanMigrate judge if current state can migrate to given toState by one or more step
func (f *FSM[T, S, U, V]) CanMigrate(toState T) bool

```

## Callbacks

### Ordinary Callbacks Usage

You can use:

```go
func (f *FSM[T, S, U, V]) SetCallbacks(Callbacks *Callbacks[T, S, U, V])
```

to set up callback functions that will be executed in the `Trigger()`.

```go
// Callbacks do something while eventE is triggering
type Callbacks[T, S comparable, U, V any] struct {
    onEntry           func(*Event[T, S, U, V]) error    // will be executed in any case
    beforeStateChange func(*Event[T, S, U, V]) error
    afterStateChange  func(*Event[T, S, U, V]) error
    onDefer           func(*Event[T, S, U, V], error)   // will be executed in any case
}
```

```mermaid
flowchart LR
    onEntry[onEntry\nwill be executed in any case]-->beforeStateChange
    beforeStateChange-->S(*FSM State Change*)
    S-->afterStateChange
    afterStateChange-->onDefer[onDefer\nwill be executed in any case]
```

`onEntry` and `onDefer` will be executed in any case and can be used for some tasks such as resource allocate/release, data statistics, etc.

A common use is to use pointer types to pass in parameters or get return values from Callbacks

### Advanced Callbacks Usage

The callback function can access the **custom attributes** of **any** `Event` and `State` when it is executed. It means that you can define custom attributes as functions to execute, and you can also integrate your callback function design in one config to avoid multiple configs.

```go
testFSM.SetCallbacks(&Callbacks[nodeState, eventVal, edgeVal, nodeVal]{
    afterStateChange: func(e *Event[nodeState, eventVal, edgeVal, nodeVal]) error {
        return e.EventE().storeVal(e, w, t) // call a custom function in event store value
    },
})
```


## Principles and Terminology

This `Finite State Machine` module is based on the data structure: `Graph`. The mapping between `FSM` and `Graph` is: `State` in `FSM` maps to `Vertex` in `Graph`, and `Event` in `FSM` maps to `Edge` in `Graph`.

### FSM & Graph

In this module, the difference between `Graph` and `FSM` is: `FSM` are **stateful**,`Graph` is **stateless**, `Graph` can be considered as a **Config** of `FSM`.

```mermaid
classDiagram

    direction LR

    FSM o-- Graph
    Graph o-- EdgeCollection
    Graph o-- Vertex
    EdgeCollection o-- Edge
    Edge o-- Vertex

    class FSM~T,S,U,V~
    FSM : *Graph[T, S, U, V] g
    FSM : T prevState
    FSM : T currState 
    FSM : *Callbacks[T, S, U, V] Callbacks

    class Graph~T,S,U,V~
    Graph : [][]*EdgeCollection[T, S, U, V] adj
    Graph : map[T]*Vertex stoV
    Graph : itoV []*Vertex[T, V] itoV  
    Graph : NextEdge(fromState T, eventName S) itoV
    
    class EdgeCollection~T,S,U,V~
    EdgeCollection : []*Edge[T, S, U, V] eList
    EdgeCollection : map[S][]*Edge[T, S, U, V] eFast
    
    class Edge~T,S,U,V~
    Edge : *Vertex[T, V] fromV
    Edge : *Vertex[T, V] toV
    Edge : S eventVal
    Edge : U storeVal
    
    class Vertex~T,V~
    Vertex : int idx
    Vertex : T stateVal
    Vertex : V storeVal
```

### Event & Edge

```go
// Edge Event value included
type Edge[T, S comparable, U, V any] struct {
    fromV    *Vertex[T, V] // From vertex
    toV      *Vertex[T, V] // To vertex
    eventVal S             // Event value. Not unique
    storeVal U             // Anything you want. e.g. Real callback function(use Callbacks to invoke)
}
```


| FSM                     | Graph            | Type            | Description                                                                                                |
|-------------------------|------------------|-----------------|------------------------------------------------------------------------------------------------------------|
| Event(abstract)         | Edge             | /               | An abstract container that stores from and to Vertex and other values                                      |
| Event's from & to state | Edge.fromV & toV | `*Vertex[T, V]` | The from and to abstract state is stored, and the state attributes can be obtained from here               |
| Event's value           | Edge.eventVal    | `S comparable`  | FSM's event value to express business meaning. E.g "take an order"                                         |
| Event's other attribute | Edge.storeVal    | `U any`         | Define your own data structure to store any other event attribute or callback functions of Event dimension |

### State & Vertex

```go
// Vertex idx start with number 0
type Vertex[T comparable, V any] struct {
	idx      int // Vertex idx. Auto generated based on unique stateVal
	stateVal T   // State value. Need to be unique
	storeVal V   // Anything you want
}
```

| FSM                     | Graph           | Type           | Description                                                                                                |
|-------------------------|-----------------|----------------|------------------------------------------------------------------------------------------------------------|
| State(abstract)         | Vertex          | /              | An abstract container that encapsulates all state properties                                               |
| /                       | Vertex.idx      | `int`          | Automatically generated according to config, used for graph, meaningless to FSM                            |
| State's value           | Vertex.stateVal | `T comparable` | FSM's unique state value to express business meaning. E.g "paid" or "2"                                    |
| State's other attribute | Vertex.storeVal | `V any  `      | Define your own data structure to store any other state attribute or callback functions of State dimension |

## Built With

* [go-generic-collection](https://github.com/kiexu/go-generic-collection) - A Java-style generic collection lib of Go
* [gin](https://github.com/gin-gonic/gin) - The Web Framework used
* [mermaid](https://github.com/mermaid-js/mermaid) - The JavaScript visualization lib
* [uuid](https://github.com/google/uuid) - To generate UUID for FSM

## Contributing

[Kie Xu](https://github.com/kiexu)

## License
MIT