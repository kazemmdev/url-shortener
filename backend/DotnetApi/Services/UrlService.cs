namespace DotnetApi.Services;

using DotnetApi.Data;
using DotnetApi.Models;
using Microsoft.EntityFrameworkCore;
using StackExchange.Redis;


public class UrlService(AppDbContext db, IConfiguration configuration, IConnectionMultiplexer redis) : IUrlService
{
    private readonly IDatabase cache = redis.GetDatabase();

    public async Task<string?> GetByShortCode(string shortCode)
    {
        var cachedUrl = await cache.StringGetAsync(shortCode);
        if (cachedUrl.HasValue)
        {
            return cachedUrl.ToString();
        }

        var url = await db.Urls.FirstOrDefaultAsync(u => u.ShortCode == shortCode);
        if (url is null || url.ExpiresAt < DateTime.UtcNow)
        {
            return null;
        }

        var expiration = url.ExpiresAt - DateTime.UtcNow;

        await cache.StringSetAsync(shortCode, url.LongUrl, expiration);

        return url.LongUrl;
    }

    public async Task<string> Create(string longUrl)
    {
        var expiration = TimeSpan.FromDays(configuration.GetValue<int>("App:ExpireUrlTime"));
        var total = await db.Urls.CountAsync();

        var url = new Url
        {
            LongUrl = longUrl,
            ShortCode = ShortcodeService.Generate(total),
            ExpiresAt = DateTime.UtcNow.Add(expiration)
        };

        db.Urls.Add(url);
        await db.SaveChangesAsync();

        return url.ShortCode;
    }
}
