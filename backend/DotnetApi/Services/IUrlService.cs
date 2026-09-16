namespace DotnetApi.Services;

public interface IUrlService
{
    Task<string> GetByShortCode(string shortCode);
    Task<string> Create(string longUrl, TimeSpan expiration);
}
