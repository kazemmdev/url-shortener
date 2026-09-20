using DotnetApi.Data;
using DotnetApi.Services;
using Microsoft.EntityFrameworkCore;
using StackExchange.Redis;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddSingleton<IConnectionMultiplexer>(ConnectionMultiplexer.Connect(
        builder.Configuration.GetConnectionString("Redis")!
));

builder.Services.AddDbContext<AppDbContext>(opt => opt.UseSqlServer(
    builder.Configuration.GetConnectionString("Default")
));

builder.Services.AddScoped<IUrlService, UrlService>();
builder.Services.AddControllers();
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

var app = builder.Build();

if (builder.Configuration.GetValue<bool>("App:ApplyMigrations"))
{
    using var scope = app.Services.CreateScope();
    scope.ServiceProvider.GetRequiredService<AppDbContext>().Database.Migrate();
}

if (builder.Configuration.GetValue<bool>("App:EnableSwagger"))
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.MapControllers();

app.Run();
