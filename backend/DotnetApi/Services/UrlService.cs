namespace DotnetApi.Services;


public class UrlService() : IUrlService
{
    public Task<string> Create(string longUrl, TimeSpan expiration)
    {
        throw new NotImplementedException();
    }

    public Task<string> GetByShortCode(string shortCode)
    {
        throw new NotImplementedException();
    }
}
