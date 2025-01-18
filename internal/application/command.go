package application


type Command struct { 
	Name 				string
	Arguments		[]string 	
	Handler  func (*App, Command)error
}
