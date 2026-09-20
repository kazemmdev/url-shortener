namespace DotnetApi.Services;

using DotnetApi.Data;
using DotnetApi.Models;
using Microsoft.EntityFrameworkCore;

public class UrlService(AppDbContext db, IConfiguration configuration) : IUrlService
{
    public async Task<string?> GetByShortCode(string shortCode)
    {
        var url = await db.Urls.FirstOrDefaultAsync(u => u.ShortCode == shortCode);
        if (url is null || url.ExpiresAt < DateTime.UtcNow)
        {
            return null;
        }

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
