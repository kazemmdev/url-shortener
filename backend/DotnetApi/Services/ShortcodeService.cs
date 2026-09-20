namespace DotnetApi.Services;

using System.Security.Cryptography;

public static class ShortcodeService
{
    private const string Alphabet =
    "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";

    private const int ShortCodeLength = 7;
    private const int PaddedLength = 9;
    private const long Salt = 123;

    public static string Generate(long counter)
    {
        // 1. Add salt
        var value = counter + Salt;

        // 2. Convert to exactly 9 characters
        var padded = value.ToString($"D{PaddedLength}");

        // 3. Shuffle the 9 characters
        var shuffled = Shuffle(padded);

        // 4. Convert the shuffled numeric string to Base62
        var number = long.Parse(shuffled);

        return ToBase62(number);
    }

    private static string Shuffle(string value)
    {
        var chars = value.ToCharArray();

        // Fisher-Yates shuffle
        for (var i = chars.Length - 1; i > 0; i--)
        {
            var j = RandomNumberGenerator.GetInt32(i + 1);
            (chars[i], chars[j]) = (chars[j], chars[i]);
        }

        return new string(chars);
    }

    private static string ToBase62(long value)
    {
        if (value == 0)
        {
            return "0";
        }

        Span<char> buffer = stackalloc char[ShortCodeLength];
        var position = buffer.Length;

        while (value > 0)
        {
            buffer[--position] = Alphabet[(int)(value % 62)];
            value /= 62;
        }

        return new string(buffer[position..]);
    }
}