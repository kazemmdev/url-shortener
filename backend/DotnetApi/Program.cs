var builder = WebApplication.CreateBuilder(args);



var connection = builder.Configuration.GetConnectionString("Default");


var app = builder.Build();

app.Run();
