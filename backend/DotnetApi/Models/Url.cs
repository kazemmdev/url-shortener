namespace DotnetApi.Models;

using System.ComponentModel.DataAnnotations;

public class Url
{
    [Key]
    public required string ShortCode { get; set; }
    public required string LongUrl { get; set; }
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    public DateTime ExpiresAt { get; set; }
}
