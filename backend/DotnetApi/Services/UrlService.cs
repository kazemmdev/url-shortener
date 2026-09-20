namespace DotnetApi.Services;

using System.Security.Cryptography;
using DotnetApi.Data;
using DotnetApi.Models;
using Microsoft.EntityFrameworkCore;

public class UrlService(AppDbContext db, IConfiguration configuration) : IUrlService
{
    private const string Alphabet = "0123456789abcdefghijklmnopqrstuvwxyz";
    private const int ShortCodeLength = 3;

    // private const string Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
    // private const int ShortCodeLength = 7;

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

        string shortCode;
        do
        {
            shortCode = GenerateShortCode();
        } while (await db.Urls.AnyAsync(u => u.ShortCode == shortCode));

        db.Urls.Add(new Url
        {
            ShortCode = shortCode,
            LongUrl = longUrl,
            ExpiresAt = DateTime.UtcNow.Add(expiration)
        });
        await db.SaveChangesAsync();

        return shortCode;
    }

    private static string GenerateShortCode()
    {
        Span<char> chars = stackalloc char[ShortCodeLength];
        for (var i = 0; i < ShortCodeLength; i++)
        {
            chars[i] = Alphabet[RandomNumberGenerator.GetInt32(Alphabet.Length)];
        }

        return new string(chars);
    }
}
