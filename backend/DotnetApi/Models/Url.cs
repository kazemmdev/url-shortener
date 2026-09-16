namespace DotnetApi.Models;

public class Url
{   
    public string ShortCode { get; set; }
    public string LongUrl { get; set; }
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    public DateTime ExpiresAt { get; set; }
}
