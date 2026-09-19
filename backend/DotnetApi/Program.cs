using DotnetApi.Data;
using Microsoft.EntityFrameworkCore;

var builder = WebApplication.CreateBuilder(args);

var connection = builder.Configuration.GetConnectionString("Default");

builder.Services.AddDbContext<AppDbContext>(opt => opt.UseSqlServer(connection));

var app = builder.Build();

app.Run();
