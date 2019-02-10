package main

func main() {
	service := micro.NewService(
		micro.Name("service.auth"),
	)

	service.Init()

	//auth.RegisterAuthHandler(service.Server(), &Auth{
	//	customers: loadCustomerData("data/customers.json"),
	//})

	service.Run()
}