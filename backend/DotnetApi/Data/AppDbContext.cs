namespace DotnetApi.Data;

using DotnetApi.Models;
using Microsoft.EntityFrameworkCore;

public class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
{
    public DbSet<Url> Urls => Set<Url>();
}